// SPDX-License-Identifier: MIT
import "./app.css";
import { mount } from "svelte";
import { displayName } from "$lib/brand";
import App from "./App.svelte";

document.title = displayName;

const app = mount(App, {
  target: document.getElementById("app")!,
});

export default app;
