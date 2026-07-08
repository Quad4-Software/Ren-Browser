// SPDX-License-Identifier: MIT

export function portal(node: HTMLElement, target: ParentNode = document.body) {
  target.appendChild(node);
  return {
    destroy() {
      if (node.parentNode === target) {
        node.remove();
      }
    },
  };
}

export function stopPointerBubble(node: HTMLElement) {
  const stop = (event: PointerEvent) => {
    event.stopPropagation();
  };
  node.addEventListener("pointerdown", stop);
  return {
    destroy() {
      node.removeEventListener("pointerdown", stop);
    },
  };
}
