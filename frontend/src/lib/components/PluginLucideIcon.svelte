<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { Puzzle } from "@lucide/svelte";
  import { resolvePluginLucideIcon } from "$lib/plugins/plugin-icons.js";

  type Props = {
    name?: string | null;
    size?: number;
    class?: string;
  };

  let { name = null, size = 16, class: className = "" }: Props = $props();

  const Icon = $derived(resolvePluginLucideIcon(name) ?? Puzzle);
</script>

<Icon {size} class={className} aria-hidden="true" />
