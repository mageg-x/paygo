<template>
  <div class="json-editor-wrapper">
    <div class="editor-toolbar">
      <button @click="formatJson" class="toolbar-btn" title="格式化">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M4 6h16M4 12h16M4 18h12" />
        </svg>
        格式化
      </button>
      <button @click="compressJson" class="toolbar-btn" title="压缩">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M19 9l-7 7-7-7" />
        </svg>
        压缩
      </button>
      <button @click="copyJson" class="toolbar-btn" title="复制">
        <svg class="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M8 16H6a2 2 0 01-2-2V6a2 2 0 012-2h8a2 2 0 012 2v2m-6 12h8a2 2 0 002-2v-8a2 2 0 00-2-2h-8a2 2 0 00-2 2v8a2 2 0 002 2z" />
        </svg>
        复制
      </button>
      <span class="toolbar-hint">JSON 语法高亮</span>
    </div>
    <div class="editor-container">
      <div class="line-numbers" ref="lineNumbers">
        <div v-for="n in lineCount" :key="n">{{ n }}</div>
      </div>
      <textarea
        ref="textarea"
        v-model="content"
        @input="onInput"
        @scroll="syncScroll"
        :placeholder="placeholder"
        class="editor-textarea"
        spellcheck="false"
      ></textarea>
      <div class="highlight-layer" ref="highlightLayer" v-html="highlightedHtml"></div>
    </div>
    <div v-if="error" class="editor-error">{{ error }}</div>
  </div>
</template>

<script setup lang="ts">
import { ref, computed, watch, onMounted } from 'vue'
import { ElMessage } from 'element-plus'

const props = defineProps<{
  modelValue: string
  placeholder?: string
}>()

const emit = defineEmits<{
  (e: 'update:modelValue', value: string): void
  (e: 'change', value: string): void
}>()

const textarea = ref<HTMLTextAreaElement | null>(null)
const highlightLayer = ref<HTMLDivElement | null>(null)
const lineNumbers = ref<HTMLDivElement | null>(null)
const content = ref(props.modelValue || '{}')
const error = ref('')

const lineCount = computed(() => {
  return content.value.split('\n').length
})

// 简单的 JSON 语法高亮
const highlightedHtml = computed(() => {
  const text = content.value
  if (!text) return ''

  // 转义 HTML
  let html = text
    .replace(/&/g, '&amp;')
    .replace(/</g, '&lt;')
    .replace(/>/g, '&gt;')

  // 字符串（键）
  html = html.replace(/"([^"]+)":/g, '<span class="json-key">"$1"</span>:')
  // 字符串（值）
  html = html.replace(/: "([^"]*)"/g, ': <span class="json-string">"$1"</span>')
  // 数字
  html = html.replace(/: (\d+\.?\d*)/g, ': <span class="json-number">$1</span>')
  // 布尔值
  html = html.replace(/: (true|false)/g, ': <span class="json-boolean">$1</span>')
  // null
  html = html.replace(/: (null)/g, ': <span class="json-null">$1</span>')

  return html
})

function onInput() {
  // 验证 JSON
  if (content.value.trim()) {
    try {
      JSON.parse(content.value)
      error.value = ''
    } catch (e: any) {
      error.value = e.message
    }
  } else {
    error.value = ''
  }

  emit('update:modelValue', content.value)
  emit('change', content.value)
}

function formatJson() {
  try {
    const parsed = JSON.parse(content.value)
    content.value = JSON.stringify(parsed, null, 2)
    error.value = ''
  } catch (e: any) {
    error.value = 'JSON格式错误: ' + e.message
  }
}

function compressJson() {
  try {
    const parsed = JSON.parse(content.value)
    content.value = JSON.stringify(parsed)
    error.value = ''
  } catch (e: any) {
    error.value = 'JSON格式错误: ' + e.message
  }
}

function copyJson() {
  navigator.clipboard.writeText(content.value).then(() => {
    ElMessage.success('已复制到剪贴板')
  })
}

function syncScroll() {
  if (highlightLayer.value && textarea.value) {
    highlightLayer.value.scrollTop = textarea.value.scrollTop
    highlightLayer.value.scrollLeft = textarea.value.scrollLeft
  }
  if (lineNumbers.value && textarea.value) {
    lineNumbers.value.scrollTop = textarea.value.scrollTop
  }
}

watch(() => props.modelValue, (newVal) => {
  if (newVal !== content.value) {
    content.value = newVal || '{}'
  }
})

onMounted(() => {
  if (!props.modelValue) {
    content.value = '{}'
  }
})
</script>

<style scoped>
.json-editor-wrapper {
  border: 1px solid #e5e7eb;
  border-radius: 8px;
  overflow: hidden;
  background: #fafbfc;
}

.editor-toolbar {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 8px 12px;
  background: #f3f4f6;
  border-bottom: 1px solid #e5e7eb;
}

.toolbar-btn {
  display: flex;
  align-items: center;
  gap: 4px;
  padding: 4px 8px;
  font-size: 12px;
  color: #4b5563;
  background: white;
  border: 1px solid #d1d5db;
  border-radius: 4px;
  cursor: pointer;
  transition: all 0.2s;
}

.toolbar-btn:hover {
  background: #e5e7eb;
  border-color: #9ca3af;
}

.toolbar-hint {
  margin-left: auto;
  font-size: 11px;
  color: #9ca3af;
}

.editor-container {
  position: relative;
  display: flex;
  min-height: 200px;
  max-height: 400px;
  overflow: hidden;
}

.line-numbers {
  padding: 12px 8px;
  background: #f3f4f6;
  border-right: 1px solid #e5e7eb;
  text-align: right;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  color: #9ca3af;
  user-select: none;
  overflow: hidden;
  min-width: 40px;
}

.editor-textarea,
.highlight-layer {
  flex: 1;
  padding: 12px;
  font-family: 'Monaco', 'Menlo', 'Ubuntu Mono', monospace;
  font-size: 13px;
  line-height: 1.5;
  tab-size: 2;
  white-space: pre;
  overflow: auto;
}

.editor-textarea {
  position: absolute;
  top: 0;
  left: 40px;
  right: 0;
  bottom: 0;
  width: calc(100% - 40px);
  height: 100%;
  border: none;
  outline: none;
  resize: none;
  background: transparent;
  color: transparent;
  caret-color: #374151;
  z-index: 2;
}

.highlight-layer {
  pointer-events: none;
  background: white;
  color: #374151;
  z-index: 1;
}

.editor-textarea::placeholder {
  color: #9ca3af;
}

.editor-error {
  padding: 8px 12px;
  background: #fef2f2;
  border-top: 1px solid #fecaca;
  color: #dc2626;
  font-size: 12px;
}

/* JSON 语法高亮颜色 */
:deep(.json-key) { color: #7c3aed; }
:deep(.json-string) { color: #059669; }
:deep(.json-number) { color: #2563eb; }
:deep(.json-boolean) { color: #d97706; }
:deep(.json-null) { color: #6b7280; }
</style>
