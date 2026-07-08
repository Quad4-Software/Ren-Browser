<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import { createApp, type AppController } from "$lib/app/create-app.svelte";

  type Props = {
    onReady: (app: AppController) => void;
  };

  let { onReady }: Props = $props();

  const app = createApp();

  $effect(() => {
    onReady(app);
  });
</script>
