<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const props = withDefaults(
  defineProps<{
    messageKey?: string
  }>(),
  {
    messageKey: 'common.state.empty',
  },
)

const localeStore = useLocaleStore()
</script>

<template>
  <section class="empty-state" role="status">
    <p>{{ localeStore.t(props.messageKey) }}</p>
  </section>
</template>

<style scoped>
.empty-state {
  border: 1px dashed #c0c4cc;
  border-radius: 0.5rem;
  color: #606266;
  padding: 2rem;
  text-align: center;
}
</style>
