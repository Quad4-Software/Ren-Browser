// SPDX-License-Identifier: MIT
import "./app.css";
import { mount } from "svelte";
import { displayName } from "$lib/brand";
import { fetchAuthStatus } from "$lib/auth/api";
import { applyScreenshotQueryTheme } from "$lib/theme/screenshot";
import App from "./App.svelte";
import AuthApp from "./AuthApp.svelte";

document.title = displayName;
applyScreenshotQueryTheme();

async function boot() {
  const target = document.getElementById("app");
  if (!target) {
    return;
  }

  const status = await fetchAuthStatus();
  if (status.authRequired && !status.authenticated) {
    mount(AuthApp, { target });
    return;
  }

  mount(App, { target });
}

void boot();
