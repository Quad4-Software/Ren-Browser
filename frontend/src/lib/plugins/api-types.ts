// SPDX-License-Identifier: MIT
export type PluginPanelId = `plugin:${string}`;

export type ActivePanel =
  "browser" | "discovery" | "history" | "devtools" | "settings" | PluginPanelId;

export type Disposable = {
  dispose(): void;
};

export type PluginContext = {
  pluginId: string;
  subscriptions: {
    add(disposable: Disposable): void;
  };
  storage: {
    get(key: string): Promise<string | null>;
    set(key: string, value: string): Promise<void>;
  };
  navigation: {
    getCurrentURL(): string;
    navigate(url: string): void;
  };
  events: {
    on(event: string, fn: (data: unknown) => void): Disposable;
    emit(event: string, data: unknown): void;
  };
  ui: {
    showToast(message: string): void;
  };
};

export type PluginModule = {
  activate?(ctx: PluginContext): void | Promise<void>;
  deactivate?(): void | Promise<void>;
  mount?(el: HTMLElement): void | Promise<void>;
  handleScheme?(
    url: string,
  ): { html: string; contentType: string } | Promise<{ html: string; contentType: string }>;
};

export type PluginCommand = {
  pluginId: string;
  commandId: string;
  title: string;
  keybind?: string;
};

export type PluginPanelContribution = {
  pluginId: string;
  id: string;
  title: string;
  icon?: string;
  entry: string;
  location?: string;
};

export type PluginDevToolsContribution = {
  pluginId: string;
  id: string;
  title: string;
  entry: string;
};

export type ContributionsSnapshot = {
  panels: PluginPanelContribution[];
  commands: PluginCommand[];
  devtools: PluginDevToolsContribution[];
  urlSchemes: Array<{ pluginId: string; scheme: string; handler?: string }>;
};
