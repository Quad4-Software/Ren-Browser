export function activate(ctx) {
  ctx.subscriptions.add(
    ctx.events.on("ready", () => {
      ctx.ui.showToast("Hello extension activated");
    }),
  );
  ctx.events.emit("ready", {});
}

export function deactivate() {}

export function mount(el) {
  el.innerHTML =
    '<article class="hello-panel"><h2>Hello panel</h2><p>This sidebar panel is contributed by the hello extension.</p></article>';
}

export function handleScheme(url) {
  return {
    html: `<article class="hello-page"><h1>Hello from extension</h1><p>URL: ${url}</p></article>`,
    contentType: "html",
  };
}
