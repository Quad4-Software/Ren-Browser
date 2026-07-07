// SPDX-License-Identifier: MIT
import "./app.css";
import { mount } from "svelte";
import { formatBindingError } from "$lib/browser/binding-errors";
import { displayName } from "$lib/brand";
import { fetchAuthStatus } from "$lib/auth/api";
import { applyScreenshotQueryTheme } from "$lib/theme/screenshot";
import App from "./App.svelte";
import AuthApp from "./AuthApp.svelte";
import BootShell from "./BootShell.svelte";
import CrashPage from "$lib/components/CrashPage.svelte";
import { Shutdown } from "../bindings/renbrowser/internal/app/browserservice.js";

document.title = displayName;
applyScreenshotQueryTheme();

function showBootError(message: string, cause?: unknown) {
  const target = document.getElementById("app");
  if (!target) {
    return;
  }
  target.replaceChildren();
  mount(CrashPage, {
    target,
    props: {
      message,
      cause,
      onReload: () => location.reload(),
      onClose: async () => {
        try {
          await Shutdown();
        } catch {
          window.close();
        }
      },
    },
  });
}

async function boot() {
  const target = document.getElementById("app");
  if (!target) {
    return;
  }

  try {
    const status = await fetchAuthStatus();
    if (status.authRequired && !status.authenticated) {
      mount(BootShell, { target, props: { Root: AuthApp } });
      return;
    }

    mount(BootShell, { target, props: { Root: App } });
  } catch (err) {
    const message = formatBindingError(err, "Unexpected startup error");
    showBootError(message, err);
  }
}

void boot();
