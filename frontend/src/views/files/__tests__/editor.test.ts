import { beforeEach, describe, expect, it, vi } from "vitest";
import Editor from "../Editor.vue";

const harness = vi.hoisted(() => ({
  put: vi.fn(),
  showError: vi.fn(),
  showHover: vi.fn(),
  closeHovers: vi.fn(),
  user: { perm: { modify: true }, aceEditorTheme: "" },
  req: { name: "app.ts", content: "initial", type: "text" },
}));
vi.mock("@/api", () => ({ files: { put: harness.put } }));
vi.mock("@/stores/auth", () => ({
  useAuthStore: () => ({ user: harness.user }),
}));
vi.mock("@/stores/file", () => ({
  useFileStore: () => ({ req: harness.req }),
}));
vi.mock("@/stores/layout", () => ({
  useLayoutStore: () => ({
    showHover: harness.showHover,
    closeHovers: harness.closeHovers,
  }),
}));
vi.mock("@/utils/theme", () => ({ getEditorTheme: () => "github-light" }));
vi.mock("@/utils/buttons", () => ({
  default: { loading: vi.fn(), done: vi.fn(), success: vi.fn() },
}));
vi.mock("@/components/files/MonacoEditor.vue", () => ({ default: {} }));
vi.mock("@/components/header/HeaderBar.vue", () => ({ default: {} }));
vi.mock("@/components/header/Action.vue", () => ({ default: {} }));
vi.mock("@/components/Breadcrumbs.vue", () => ({ default: {} }));
vi.mock("vue-router", () => ({
  useRoute: () => ({ path: "/files/app.ts", query: {} }),
  useRouter: () => ({ push: vi.fn() }),
  onBeforeRouteLeave: vi.fn(),
  onBeforeRouteUpdate: vi.fn(),
}));
vi.mock("vue-i18n", () => ({ useI18n: () => ({ t: (key: string) => key }) }));
vi.mock("vue", async (original) => ({
  ...(await original<typeof import("vue")>()),
  inject: () => harness.showError,
  onMounted: vi.fn(),
  onBeforeUnmount: vi.fn(),
  useSSRContext: () => ({ modules: new Set<string>() }),
}));

function setup() {
  const bindings = (Editor as any).setup({}, { expose: vi.fn() });
  let version = 1;
  let text = "initial";
  let changed = () => {};
  const model = {
    getAlternativeVersionId: () => version,
    getValue: () => text,
  };
  bindings.editorReady({
    getModel: () => model,
    onDidChangeModelContent: (callback: () => void) => {
      changed = callback;
    },
    onDidChangeCursorSelection: vi.fn(),
    focus: vi.fn(),
  });
  return {
    bindings,
    change: (value: string, id: number) => {
      text = value;
      version = id;
      changed();
    },
  };
}

beforeEach(() => {
  vi.clearAllMocks();
  harness.put.mockResolvedValue({});
  harness.user.perm.modify = true;
  harness.req.type = "text";
  vi.stubGlobal("localStorage", { getItem: () => null, setItem: vi.fn() });
});

describe("Monaco saves and navigation", () => {
  it("keeps edits made during an in-flight save dirty, and recognizes undo", async () => {
    let finish!: () => void;
    harness.put.mockImplementation(
      () =>
        new Promise<void>((resolve) => {
          finish = resolve;
        })
    );
    const { bindings, change } = setup();
    expect(bindings.confirmNavigation()).toBe(true);
    change("first edit", 2);
    const saved = bindings.save();
    change("later edit", 3);
    finish();
    await saved;
    expect(harness.put).toHaveBeenCalledWith("/files/app.ts", "first edit");
    expect(bindings.isDirty.value).toBe(true);
    change("first edit", 2);
    expect(bindings.isDirty.value).toBe(false);
  });

  it("stays on the file when save fails and resolves cancellation", async () => {
    harness.put.mockRejectedValue(new Error("Save failed"));
    const { bindings, change } = setup();
    change("changed", 2);
    const navigation = bindings.confirmNavigation();
    const prompt = harness.showHover.mock.calls[0][0];
    await prompt.saveAction();
    expect(bindings.isDirty.value).toBe(true);
    expect(harness.closeHovers).not.toHaveBeenCalled();
    expect(harness.showError).toHaveBeenCalled();
    await prompt.close();
    expect(await navigation).toBe(false);
  });

  it("allows explicit discard without saving", async () => {
    const { bindings, change } = setup();
    change("changed", 2);
    const navigation = bindings.confirmNavigation();
    harness.showHover.mock.calls[0][0].confirm({ preventDefault: vi.fn() });
    expect(await navigation).toBe(true);
    expect(harness.put).not.toHaveBeenCalled();
  });

  it.each(["permission", "immutable"])(
    "does not save a read-only file (%s)",
    async (reason) => {
      if (reason === "permission") harness.user.perm.modify = false;
      else harness.req.type = "textImmutable";
      const { bindings } = setup();
      expect(await bindings.save()).toBe(false);
      expect(harness.put).not.toHaveBeenCalled();
    }
  );
});
