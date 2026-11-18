package fbhttp

import (
	"crypto/rand"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"os"
	"path"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/filebrowser/filebrowser/v2/remote"
	"github.com/gorilla/mux"
)

var remoteSettingsMu sync.Mutex

var remoteServersHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	remoteSettingsMu.Lock()
	defer remoteSettingsMu.Unlock()
	set, err := d.store.Settings.Get()
	if err != nil {
		return http.StatusInternalServerError, err
	}
	switch r.Method {
	case http.MethodGet:
		servers := append([]remote.Server{}, set.RemoteServers...)
		for i := range servers {
			servers[i].Secret = ""
		}
		w.Header().Set("Cache-Control", "no-store")
		return renderJSON(w, r, servers)
	case http.MethodPost:
		var req struct {
			remote.Server
			Credential string `json:"credential"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
			return http.StatusBadRequest, err
		}
		req.ID = rand.Text()
		if err := req.Validate(); err != nil {
			return http.StatusBadRequest, err
		}
		if err := req.SetSecret(req.Credential, set.Key); err != nil {
			return http.StatusBadRequest, err
		}
		client, closeClient, err := req.Connect(r.Context(), set.Key)
		if err != nil {
			return http.StatusBadGateway, err
		}
		defer closeClient()
		root, err := remote.Resolve(client, req.Root, "/")
		if err != nil {
			return http.StatusBadRequest, err
		}
		info, err := client.Stat(root)
		if err != nil {
			return http.StatusBadGateway, err
		}
		if !info.IsDir() {
			return http.StatusBadRequest, errors.New("storage root must be a directory")
		}
		set.RemoteServers = append(set.RemoteServers, req.Server)
		if err := d.store.Settings.Save(set); err != nil {
			return http.StatusInternalServerError, err
		}
		req.Secret = ""
		return renderJSON(w, r, req.Server)
	case http.MethodDelete:
		id := mux.Vars(r)["id"]
		for i, server := range set.RemoteServers {
			if server.ID == id {
				set.RemoteServers = append(set.RemoteServers[:i], set.RemoteServers[i+1:]...)
				err := d.store.Settings.Save(set)
				return errToStatus(err), err
			}
		}
		return http.StatusNotFound, nil
	}
	return http.StatusMethodNotAllowed, nil
})

type remoteEntry struct {
	Name     string    `json:"name"`
	IsDir    bool      `json:"isDir"`
	Size     int64     `json:"size"`
	Modified time.Time `json:"modified"`
}

var remoteFilesHandler = withAdmin(func(w http.ResponseWriter, r *http.Request, d *data) (int, error) {
	var server *remote.Server
	for i := range d.settings.RemoteServers {
		if d.settings.RemoteServers[i].ID == mux.Vars(r)["id"] {
			server = &d.settings.RemoteServers[i]
			break
		}
	}
	if server == nil {
		return http.StatusNotFound, nil
	}
	if r.Method == http.MethodPost && !d.user.Perm.Create {
		return http.StatusForbidden, nil
	}
	if r.URL.Query().Get("download") == "true" && !d.user.Perm.Download {
		return http.StatusForbidden, nil
	}
	client, closeClient, err := server.Connect(r.Context(), d.settings.Key)
	if err != nil {
		return http.StatusBadGateway, err
	}
	defer closeClient()
	target, err := remote.Resolve(client, server.Root, r.URL.Query().Get("path"))
	if err != nil {
		return remoteStatus(err), err
	}
	w.Header().Set("Cache-Control", "no-store")
	if r.Method == http.MethodPost {
		name := r.URL.Query().Get("name")
		if name == "" || name == "." || name == ".." || path.Base(name) != name || strings.ContainsAny(name, "\\\x00") {
			return http.StatusBadRequest, errors.New("invalid filename")
		}
		destination := path.Join(target, name)
		// O_EXCL prevents overwrites and refuses existing symlinks atomically.
		file, err := client.OpenFile(destination, os.O_WRONLY|os.O_CREATE|os.O_EXCL)
		if err != nil {
			if _, statErr := client.Lstat(destination); statErr == nil {
				return http.StatusConflict, errors.New("file already exists")
			}
			return remoteStatus(err), err
		}
		_, copyErr := io.Copy(file, r.Body)
		closeErr := file.Close()
		if copyErr != nil || closeErr != nil {
			_ = client.Remove(destination)
			return http.StatusBadGateway, errors.Join(copyErr, closeErr)
		}
		return http.StatusCreated, nil
	}
	if r.URL.Query().Get("download") == "true" {
		file, err := client.Open(target)
		if err != nil {
			return remoteStatus(err), err
		}
		defer file.Close()
		info, err := file.Stat()
		if err != nil {
			return remoteStatus(err), err
		}
		if !info.Mode().IsRegular() {
			return http.StatusBadRequest, errors.New("only regular files can be downloaded")
		}
		w.Header().Set("Content-Disposition", mime.FormatMediaType("attachment", map[string]string{"filename": path.Base(target)}))
		w.Header().Set("Content-Type", "application/octet-stream")
		http.ServeContent(w, r, info.Name(), info.ModTime(), file)
		return 0, nil
	}
	entries, err := client.ReadDir(target)
	if err != nil {
		return remoteStatus(err), err
	}
	items := make([]remoteEntry, 0, len(entries))
	for _, entry := range entries {
		// Do not dereference links in listings, including links outside the root.
		if entry.Mode()&os.ModeSymlink != 0 || (!entry.IsDir() && !entry.Mode().IsRegular()) {
			continue
		}
		items = append(items, remoteEntry{Name: entry.Name(), IsDir: entry.IsDir(), Size: entry.Size(), Modified: entry.ModTime()})
	}
	sort.Slice(items, func(i, j int) bool {
		if items[i].IsDir != items[j].IsDir {
			return items[i].IsDir
		}
		return items[i].Name < items[j].Name
	})
	return renderJSON(w, r, items)
})

func remoteStatus(err error) int {
	switch {
	case errors.Is(err, os.ErrPermission):
		return http.StatusForbidden
	case errors.Is(err, os.ErrNotExist):
		return http.StatusNotFound
	default:
		return http.StatusBadGateway
	}
}
