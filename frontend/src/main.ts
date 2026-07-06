// SPDX-License-Identifier: MIT
import "./app.css";
import { mount } from "svelte";
import { displayName } from "$lib/brand";
import { fetchAuthStatus } from "$lib/auth/api";
import { applyScreenshotQueryTheme } from "$lib/theme/screenshot";
import App from "./App.svelte";
import AuthApp from "./AuthApp.svelte";
import BootShell from "./BootShell.svelte";

document.title = displayName;
applyScreenshotQueryTheme();

function showBootError(message: string) {
  const target = document.getElementById("app");
  if (!target) {
    return;
  }
  target.replaceChildren();
  const heading = document.createElement("h1");
  heading.textContent = "Ren Browser could not start";
  const detail = document.createElement("p");
  detail.textContent = message;
  target.append(heading, detail);
  Object.assign(target.style, {
    padding: "24px",
    color: "#f3f4f6",
    fontFamily: "system-ui, sans-serif",
    lineHeight: "1.5",
  });
  heading.style.fontSize = "1.1rem";
  heading.style.margin = "0 0 12px";
  detail.style.margin = "0";
  detail.style.color = "#9ca3af";
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
    const message = err instanceof Error ? err.message : "Unexpected startup error";
    showBootError(message);
  }
}

void boot();
