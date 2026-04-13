<template>
	<div class="plain-markdown-editor">
		<textarea
			v-if="isEditing"
			ref="textareaRef"
			v-model="localContent"
			class="textarea"
			:placeholder="placeholder"
			:disabled="!isEditEnabled"
			@input="onInput"
		/>
		<!-- eslint-disable-next-line vue/no-v-html -->
		<div
			v-else-if="renderedContent"
			class="content"
			v-html="renderedContent"
		/>
		<p
			v-else
			class="has-text-grey"
		>
			{{ placeholder }}
		</p>

		<ul
			v-if="!isEditing && isEditEnabled"
			class="tiptap__editor-actions d-print-none"
		>
			<li>
				<BaseButton
					class="done-edit"
					@click="startEditing"
				>
					{{ $t('input.editor.edit') }}
				</BaseButton>
			</li>
		</ul>
		<XButton
			v-else-if="isEditing && showSave"
			v-cy="'saveEditor'"
			class="mbs-4"
			variant="secondary"
			:shadow="false"
			:disabled="!contentHasChanged"
			@click="bubbleSave"
		>
			{{ $t('misc.save') }}
		</XButton>
	</div>
</template>

<script setup lang="ts">
import {ref, computed, watch, nextTick} from 'vue'
import {marked} from 'marked'
import BaseButton from '@/components/base/BaseButton.vue'
import XButton from '@/components/input/Button.vue'
import {saveEditorDraft, loadEditorDraft, clearEditorDraft} from '@/helpers/editorDraftStorage'

const props = withDefaults(defineProps<{
	isEditEnabled?: boolean,
	showSave?: boolean,
	placeholder?: string,
	storageKey?: string,
}>(), {
	isEditEnabled: true,
	showSave: false,
	placeholder: '',
	storageKey: '',
})

const emit = defineEmits(['save'])

const modelValue = defineModel<string>({default: ''})

const isEditing = ref(false)
const localContent = ref('')
const textareaRef = ref<HTMLTextAreaElement | null>(null)

// Track original content to detect changes
const originalContent = ref('')
const contentHasChanged = computed(() => localContent.value !== originalContent.value)

const renderedContent = computed(() => {
	const content = modelValue.value
	if (!content || content.trim() === '') {
		return ''
	}
	return marked.parse(content) as string
})

watch(() => modelValue.value, (newVal) => {
	if (!isEditing.value) {
		localContent.value = newVal
	}
}, {immediate: true})

function startEditing() {
	if (!props.isEditEnabled) {
		return
	}

	// Load draft from localStorage if available
	const draft = props.storageKey ? loadEditorDraft(props.storageKey) : null
	localContent.value = draft ?? modelValue.value
	originalContent.value = modelValue.value
	isEditing.value = true

	nextTick(() => {
		textareaRef.value?.focus()
	})
}

function onInput() {
	modelValue.value = localContent.value

	if (props.storageKey) {
		saveEditorDraft(props.storageKey, localContent.value)
	}
}

function bubbleSave() {
	modelValue.value = localContent.value
	originalContent.value = localContent.value

	if (props.storageKey) {
		clearEditorDraft(props.storageKey)
	}

	emit('save', localContent.value)
	isEditing.value = false
}
</script>

<style lang="scss" scoped>
.plain-markdown-editor {
	.textarea {
		width: 100%;
		min-height: 10rem;
		font-family: monospace;
		font-size: .9rem;
		resize: vertical;
	}
}

ul.tiptap__editor-actions {
	font-size: .8rem;
	margin: 0;

	li {
		display: inline-block;

		&::after {
			content: '·';
			padding: 0 .25rem;
		}

		&:last-child:after {
			content: '';
		}
	}

	&, a {
		color: var(--grey-500);

		&.done-edit {
			color: var(--primary);
		}
	}
}
</style>
