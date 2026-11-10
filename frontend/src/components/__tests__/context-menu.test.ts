import { describe, expect, it, vi } from "vitest";
import { reactive, ref } from "vue";
import ContextMenu from "../ContextMenu.vue";

vi.mock("@vueuse/core", () => ({
  useElementBounding: () => ({ width: ref(200), height: ref(300) }),
  useWindowSize: () => ({ width: ref(800), height: ref(600) }),
}));
vi.mock("vue", async (original) => ({
  ...(await original<typeof import("vue")>()),
  watch: vi.fn(),
  onUnmounted: vi.fn(),
  useSSRContext: () => ({ modules: new Set<string>() }),
}));

function setup(x: number, y: number) {
  const props = reactive({ show: true, pos: { x, y } });
  const bindings = (ContextMenu as any).setup(props, {
    expose: vi.fn(),
    emit: vi.fn(),
  });
  return { props, ...bindings };
}

describe("context menu positioning", () => {
  it("opens below the pointer when there is room", () => {
    expect(setup(100, 80).top.value).toBe(80);
  });
  it("anchors its bottom at the pointer when the menu would overflow below", () => {
    expect(setup(100, 560).top.value).toBe(260);
  });
  it("repositions an already open menu and clamps both viewport edges", () => {
    const menu = setup(790, 560);
    expect(menu.left.value).toBe(592);
    menu.props.pos = { x: -10, y: 80 };
    expect(menu.top.value).toBe(80);
    expect(menu.left.value).toBe(8);
    menu.props.pos.y = 299;
    expect(menu.top.value).toBe(8);
  });
});
