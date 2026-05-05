<script setup lang="ts" generic="T extends Record<string, unknown>">
import { useLocaleStore } from '../stores/locale'

export interface AppTableColumn {
  prop: string
  labelKey: string
  label?: string
  width?: string | number
  minWidth?: string | number
  sortable?: boolean | 'custom'
  align?: 'left' | 'center' | 'right'
}

const props = withDefaults(
  defineProps<{
    columns: AppTableColumn[]
    data: T[]
    loading?: boolean
    total?: number
    page?: number
    pageSize?: number
    pageSizes?: number[]
    showPagination?: boolean
    emptyMessageKey?: string
  }>(),
  {
    loading: false,
    total: 0,
    page: 1,
    pageSize: 10,
    pageSizes: () => [10, 20, 50, 100],
    showPagination: true,
    emptyMessageKey: 'common.state.empty',
  },
)

const emit = defineEmits<{
  'update:page': [page: number]
  'update:pageSize': [size: number]
  'sort-change': [sort: { prop: string; order: string | null }]
  'row-click': [row: T, column: unknown, event: MouseEvent]
}>()

const localeStore = useLocaleStore()
</script>

<template>
  <div class="app-table">
    <el-table
      :data="props.data"
      v-loading="props.loading"
      stripe
      border
      style="width: 100%"
      @sort-change="emit('sort-change', $event)"
      @row-click="(row, column, event) => emit('row-click', row as T, column, event)"
    >
      <template #empty>
        <el-empty :description="localeStore.t(props.emptyMessageKey!)" />
      </template>
      <el-table-column
        v-for="col in props.columns"
        :key="col.prop"
        :prop="col.prop"
        :label="col.label ?? localeStore.t(col.labelKey)"
        :width="col.width"
        :min-width="col.minWidth"
        :sortable="col.sortable"
        :align="col.align ?? 'left'"
        show-overflow-tooltip
      />
    </el-table>
    <div v-if="props.showPagination && props.total > 0" class="pagination-wrap">
      <el-pagination
        :current-page="props.page"
        :page-size="props.pageSize"
        :page-sizes="props.pageSizes"
        :total="props.total"
        layout="total, sizes, prev, pager, next, jumper"
        background
        @current-change="(p: number) => emit('update:page', p)"
        @size-change="(s: number) => emit('update:pageSize', s)"
      />
    </div>
  </div>
</template>

<style scoped>
.app-table {
  display: grid;
  gap: 1rem;
}
.pagination-wrap {
  display: flex;
  justify-content: flex-end;
}
</style>
