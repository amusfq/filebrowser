# SSH remote storage

Administrators can open **Remote storage** under **Locations** and choose
**Add SSH storage**. Supply the SSH hostname, port, username, an absolute storage
folder, and either a password or an unencrypted OpenSSH/PEM private key. The
server must support SFTP. The connection and folder are checked before saving.

Obtain the SSH host-key SHA256 fingerprint from the server administrator over a
trusted channel. On the server, for example:

```sh
ssh-keygen -lf /etc/ssh/ssh_host_ed25519_key.pub -E sha256
```

Copy the `SHA256:...` value. Connections reject any other host key. If a server
offers several host keys, use the fingerprint of the key negotiated with this
client. Connection errors are recorded in the File Browser server log.

Select a saved connection to browse folders, search filenames in the current
folder, download individual files, or upload new files. The selected connection,
folder, search, and page are stored in the URL and restored on refresh and browser
Back/Forward. Uploads require create permission, downloads require download
permission, and existing files are never overwritten. Removing a connection
only removes its configuration; it does not delete remote files.

Connections are shared by File Browser administrators. Other users cannot access
them. Credentials are encrypted with AES-GCM using a key derived from the
instance signing key. Protect the database and configuration backups: the
instance key is also stored there. Changing that key requires adding the
connections again.

This version provides a separate remote storage browser. Local file editing,
public shares, command hooks, recursive search, and resumable uploads do not
operate on remote storage. Symlinks are omitted from listings. Each request opens
its own SSH connection, with a 30-minute connection deadline for file transfers.
Failed uploads attempt to remove the partial file; check for partial files if
the SSH connection was lost.

Use an SSH account restricted on the server to the intended storage folder
(for example, a chrooted SFTP account). File Browser checks paths and resolves
symlinks before access, but SFTP cannot make that check atomic with opening a
file if another process changes an ancestor directory concurrently.
