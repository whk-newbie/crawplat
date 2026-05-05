<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'
import type { FormItemRule } from 'element-plus'

export interface AppFormField {
  prop: string
  labelKey: string
  label?: string
  placeholderKey?: string
  placeholder?: string
  type?: 'input' | 'textarea' | 'number' | 'select' | 'switch' | 'date'
  required?: boolean
  rules?: FormItemRule[]
  options?: { label: string; value: string | number }[]
  disabled?: boolean
  rows?: number
}

const props = withDefaults(
  defineProps<{
    fields: AppFormField[]
    modelValue: Record<string, unknown>
    loading?: boolean
    submitTextKey?: string
    cancelTextKey?: string
    showCancel?: boolean
    labelWidth?: string
  }>(),
  {
    loading: false,
    submitTextKey: 'common.actions.confirm',
    cancelTextKey: 'common.actions.cancel',
    showCancel: false,
    labelWidth: '120px',
  },
)

const emit = defineEmits<{
  'update:modelValue': [value: Record<string, unknown>]
  submit: []
  cancel: []
}>()

const localeStore = useLocaleStore()
</script>

<template>
  <el-form
    :model="props.modelValue"
    :label-width="props.labelWidth"
    @submit.prevent="emit('submit')"
  >
    <el-form-item
      v-for="field in props.fields"
      :key="field.prop"
      :prop="field.prop"
      :label="field.label ?? localeStore.t(field.labelKey)"
      :required="field.required"
      :rules="field.rules"
    >
      <el-input
        v-if="!field.type || field.type === 'input'"
        :model-value="props.modelValue[field.prop]"
        :placeholder="field.placeholder ?? (field.placeholderKey ? localeStore.t(field.placeholderKey) : '')"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
      <el-input
        v-else-if="field.type === 'textarea'"
        type="textarea"
        :rows="field.rows ?? 3"
        :model-value="props.modelValue[field.prop]"
        :placeholder="field.placeholder ?? (field.placeholderKey ? localeStore.t(field.placeholderKey) : '')"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
      <el-input-number
        v-else-if="field.type === 'number'"
        :model-value="props.modelValue[field.prop] as number"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
      <el-select
        v-else-if="field.type === 'select'"
        :model-value="props.modelValue[field.prop]"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      >
        <el-option
          v-for="opt in field.options"
          :key="opt.value"
          :label="opt.label"
          :value="opt.value"
        />
      </el-select>
      <el-switch
        v-else-if="field.type === 'switch'"
        :model-value="props.modelValue[field.prop] as boolean"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
      <el-date-picker
        v-else-if="field.type === 'date'"
        :model-value="props.modelValue[field.prop]"
        :disabled="field.disabled"
        @update:model-value="emit('update:modelValue', { ...props.modelValue, [field.prop]: $event })"
      />
    </el-form-item>
    <el-form-item>
      <el-button type="primary" :loading="props.loading" native-type="submit">
        {{ localeStore.t(props.submitTextKey!) }}
      </el-button>
      <el-button v-if="props.showCancel" @click="emit('cancel')">
        {{ localeStore.t(props.cancelTextKey!) }}
      </el-button>
    </el-form-item>
  </el-form>
</template>
