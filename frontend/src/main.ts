// SPDX-License-Identifier: MIT
import "./app.css";
import { mount } from "svelte";
import { displayName } from "$lib/brand";
import { applyScreenshotQueryTheme } from "$lib/theme/screenshot";
import App from "./App.svelte";

document.title = displayName;
applyScreenshotQueryTheme();

const app = mount(App, {
  target: document.getElementById("app")!,
});

export default app;
