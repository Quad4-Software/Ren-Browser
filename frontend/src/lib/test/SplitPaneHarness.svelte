<!-- SPDX-License-Identifier: MIT -->
<script lang="ts">
  import SplitPane from "$lib/components/SplitPane.svelte";

  type Props = {
    ratio?: number;
    onRatioChange?: (ratio: number) => void;
  };

  let { ratio = 40, onRatioChange = () => {} }: Props = $props();
</script>

{#snippet primary()}
  <div class="primary-slot">primary</div>
{/snippet}

{#snippet secondary()}
  <div class="secondary-slot">secondary</div>
{/snippet}

<SplitPane {ratio} {onRatioChange} {primary} {secondary} />
