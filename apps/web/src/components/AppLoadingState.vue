<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const props = withDefaults(
  defineProps<{
    mode?: 'skeleton' | 'text'
    loading?: boolean
    messageKey?: string
    rows?: number
  }>(),
  {
    mode: 'skeleton',
    loading: true,
    messageKey: 'common.state.loading',
    rows: 3,
  },
)

const localeStore = useLocaleStore()
</script>

<template>
  <div v-if="props.loading" v-loading="props.mode === 'text'" class="loading-state">
    <template v-if="props.mode === 'skeleton'">
      <el-skeleton :rows="props.rows" animated />
    </template>
    <template v-else>
      <span>{{ localeStore.t(props.messageKey) }}</span>
    </template>
  </div>
</template>

<style scoped>
.loading-state {
  padding: 1rem;
}
</style>
