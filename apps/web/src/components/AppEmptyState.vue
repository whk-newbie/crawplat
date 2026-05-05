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
  <el-empty :description="localeStore.t(props.messageKey)" />
</template>
