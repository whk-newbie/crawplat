<script setup lang="ts">
import { useLocaleStore } from '../stores/locale'

const props = withDefaults(
  defineProps<{
    messageKey?: string
    message?: string
    retryTextKey?: string
    retryText?: string
    showRetry?: boolean
  }>(),
  {
    messageKey: 'common.error.default',
    showRetry: true,
  },
)

const emit = defineEmits<{
  retry: []
}>()

const localeStore = useLocaleStore()
</script>

<template>
  <el-result icon="error">
    <template #title>
      {{ props.message ?? localeStore.t(props.messageKey!) }}
    </template>
    <template v-if="props.showRetry" #extra>
      <el-button type="primary" @click="emit('retry')">
        {{
          props.retryText ??
            (props.retryTextKey ? localeStore.t(props.retryTextKey) : localeStore.t('common.actions.retry'))
        }}
      </el-button>
    </template>
  </el-result>
</template>
