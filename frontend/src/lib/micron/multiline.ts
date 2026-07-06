// SPDX-License-Identifier: MIT
import BaseMicronParser from "micron-parser";

export type MicronMultilineCallbacks = {
  onArmed?: () => void;
  onDisarmed?: () => void;
  onExpanded?: (element: HTMLElement) => void;
};

export type MicronMultilineHandle = {
  teardown: () => void;
};

export function attachMicronMultilineExpansion(
  container: HTMLElement,
  callbacks: MicronMultilineCallbacks = {},
): MicronMultilineHandle {
  const onArmed = (event: Event) => {
    const element = (event as CustomEvent<{ element?: HTMLElement }>).detail?.element;
    element?.classList?.add("Mu-armed");
    callbacks.onArmed?.();
  };
  const onDisarmed = (event: Event) => {
    const element = (event as CustomEvent<{ element?: HTMLElement }>).detail?.element;
    element?.classList?.remove("Mu-armed");
    callbacks.onDisarmed?.();
  };
  const onExpanded = (event: Event) => {
    const element = (event as CustomEvent<{ element?: HTMLElement }>).detail?.element;
    if (element) {
      element.classList.add("Mu-multiline");
      callbacks.onExpanded?.(element);
    }
  };

  container.addEventListener("micron-multiline-armed", onArmed);
  container.addEventListener("micron-multiline-disarmed", onDisarmed);
  container.addEventListener("micron-field-multiline-enabled", onExpanded);

  const detachParser = BaseMicronParser.enableDoubleEnterMultiline(container, {
    windowMs: 500,
    rows: 4,
  });

  return {
    teardown: () => {
      container.removeEventListener("micron-multiline-armed", onArmed);
      container.removeEventListener("micron-multiline-disarmed", onDisarmed);
      container.removeEventListener("micron-field-multiline-enabled", onExpanded);
      detachParser();
    },
  };
}
