import { flushSync, mount, tick, unmount, type Component, type ComponentProps } from "svelte";

// eslint-disable-next-line @typescript-eslint/no-explicit-any -- test helper accepts arbitrary Svelte components
export async function mountInBody<T extends Component<any>>(
  component: T,
  props: ComponentProps<T>,
): Promise<ReturnType<typeof mount>> {
  const host = document.createElement("div");
  document.body.appendChild(host);
  let instance: ReturnType<typeof mount>;
  flushSync(() => {
    instance = mount(component, { target: host, props });
  });
  await tick();
  return instance!;
}

export function cleanupMount(instance: ReturnType<typeof mount> | null | undefined): void {
  if (instance) {
    unmount(instance);
  }
  document.body.innerHTML = "";
}
