<template>
	<PlainMarkdownEditor
		v-if="useMarkdownEditor"
		v-model="modelValue"
		:is-edit-enabled="isEditEnabled"
		:show-save="showSave"
		:placeholder="placeholder"
		:storage-key="storageKey"
		@save="emit('save', $event)"
	/>
	<TipTap
		v-else
		v-model="modelValue"
		:upload-callback="uploadCallback"
		:is-edit-enabled="isEditEnabled"
		:bottom-actions="bottomActions"
		:show-save="showSave"
		:placeholder="placeholder"
		:edit-shortcut="editShortcut"
		:enable-discard-shortcut="enableDiscardShortcut"
		:enable-mentions="enableMentions"
		:mention-project-id="mentionProjectId"
		:storage-key="storageKey"
		@save="emit('save', $event)"
	/>
</template>

<script setup lang="ts">
import {computed} from 'vue'
import {useAuthStore} from '@/stores/auth'
import {createAsyncComponent} from '@/helpers/createAsyncComponent'
import PlainMarkdownEditor from './PlainMarkdownEditor.vue'
import type {BottomAction, UploadCallback} from './types'

const TipTap = createAsyncComponent(() => import('@/components/input/editor/TipTap.vue'))

const authStore = useAuthStore()
const useMarkdownEditor = computed(() => authStore.settings.frontendSettings.useMarkdownEditor ?? false)

const modelValue = defineModel<string>({default: ''})

withDefaults(defineProps<{
	uploadCallback?: UploadCallback,
	isEditEnabled?: boolean,
	bottomActions?: BottomAction[],
	showSave?: boolean,
	placeholder?: string,
	editShortcut?: string,
	enableDiscardShortcut?: boolean,
	enableMentions?: boolean,
	mentionProjectId?: number,
	storageKey?: string,
}>(), {
	uploadCallback: undefined,
	isEditEnabled: true,
	bottomActions: () => [],
	showSave: false,
	placeholder: '',
	editShortcut: '',
	enableDiscardShortcut: false,
	enableMentions: false,
	mentionProjectId: 0,
	storageKey: '',
})

const emit = defineEmits(['save'])
</script>
