<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

export interface BreadcrumbItem {
  to?: string
  labelKey: string
  label?: string
}

defineProps<{
  items: BreadcrumbItem[]
}>()

const localeStore = useLocaleStore()
</script>

<template>
  <el-breadcrumb separator="/">
    <el-breadcrumb-item v-for="(item, index) in items" :key="index" :to="item.to">
      {{ item.label ?? localeStore.t(item.labelKey) }}
    </el-breadcrumb-item>
  </el-breadcrumb>
</template>
