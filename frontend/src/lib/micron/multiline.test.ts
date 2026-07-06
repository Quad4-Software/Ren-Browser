// SPDX-License-Identifier: MIT
import { afterEach, describe, expect, it } from "vitest";
import { attachMicronMultilineExpansion } from "./multiline";

describe("attachMicronMultilineExpansion", () => {
  let container: HTMLDivElement;
  let handle: ReturnType<typeof attachMicronMultilineExpansion> | undefined;

  afterEach(() => {
    handle?.teardown();
    handle = undefined;
    container?.remove();
  });

  it("upgrades a text input to a textarea after two Enter presses", () => {
    container = document.createElement("div");
    const input = document.createElement("input");
    input.type = "text";
    input.name = "message";
    input.value = "hello";
    container.appendChild(input);
    document.body.appendChild(container);

    let armed = false;
    let expanded = false;
    handle = attachMicronMultilineExpansion(container, {
      onArmed: () => {
        armed = true;
      },
      onExpanded: () => {
        expanded = true;
      },
    });

    input.focus();
    input.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
    );
    expect(armed).toBe(true);
    expect(container.querySelector("input")).not.toBeNull();

    input.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
    );

    const textarea = container.querySelector("textarea");
    expect(textarea).not.toBeNull();
    expect(textarea?.name).toBe("message");
    expect(textarea?.value).toBe("hello\n");
    expect(textarea?.classList.contains("Mu-multiline")).toBe(true);
    expect(expanded).toBe(true);
  });

  it("does not upgrade password inputs", () => {
    container = document.createElement("div");
    const input = document.createElement("input");
    input.type = "password";
    input.name = "secret";
    container.appendChild(input);
    document.body.appendChild(container);

    handle = attachMicronMultilineExpansion(container);

    input.focus();
    input.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
    );
    input.dispatchEvent(
      new KeyboardEvent("keydown", { key: "Enter", bubbles: true, cancelable: true }),
    );

    expect(container.querySelector("textarea")).toBeNull();
    expect(container.querySelector('input[type="password"]')).not.toBeNull();
  });
});
