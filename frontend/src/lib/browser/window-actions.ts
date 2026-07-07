// SPDX-License-Identifier: MIT
import { Window } from "@wailsio/runtime";
import { formatBindingError } from "./binding-errors.js";

export type WindowRuntime = {
  IsMaximised: () => Promise<boolean>;
  IsMinimised: () => Promise<boolean>;
  Maximise: () => Promise<void>;
  UnMaximise: () => Promise<void>;
  UnMinimise: () => Promise<void>;
};

export type TitlebarWindowAction = "maximized" | "restored";

export type TitlebarWindowActionResult = {
  ok: boolean;
  action?: TitlebarWindowAction;
  error?: string;
};

export async function handleTitlebarDoubleClick(
  windowApi: WindowRuntime = Window,
): Promise<TitlebarWindowActionResult> {
  try {
    const [maximized, minimized] = await Promise.all([
      windowApi.IsMaximised(),
      windowApi.IsMinimised(),
    ]);

    if (minimized) {
      await windowApi.UnMinimise();
      await windowApi.Maximise();
      return { ok: true, action: "maximized" };
    }

    if (maximized) {
      await windowApi.UnMaximise();
      return { ok: true, action: "restored" };
    }

    await windowApi.Maximise();
    return { ok: true, action: "maximized" };
  } catch (err) {
    return {
      ok: false,
      error: formatBindingError(err, "Could not change window size"),
    };
  }
}
