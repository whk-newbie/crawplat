<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const props = withDefaults(
  defineProps<{
    messageKey?: string
  }>(),
  {
    messageKey: 'common.state.loading',
  },
)

const localeStore = useLocaleStore()
</script>

<template>
  <section class="loading-state" role="status" aria-live="polite">
    {{ localeStore.t(props.messageKey) }}
  </section>
</template>

<style scoped>
.loading-state {
  color: #606266;
  padding: 1rem;
}
</style>
