import { describe, expect, it } from "vitest";
import { matchKeybind, parseChord } from "./keybinds";

describe("keybinds", () => {
  it("matches mod+l", () => {
    const event = {
      key: "l",
      ctrlKey: true,
      metaKey: false,
      shiftKey: false,
      altKey: false,
    } as KeyboardEvent;
    expect(matchKeybind(event, "mod+l")).toBe(true);
  });

  it("parses shift modifier", () => {
    expect(parseChord("mod+shift+i")).toEqual({
      mod: true,
      shift: true,
      alt: false,
      key: "i",
    });
  });
});
