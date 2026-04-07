<template>
  <NModal :show="show" @update:show="$emit('update:show', $event)">
    <NCard
      closable
      :mask-closable="false"
      title="裁剪头像"
      style="width: 40vh; height: 50vh"
      @close="$emit('update:show', false)"
    >
      <NFlex vertical>
        <div style="width: 100%; height: 80%">
          <VueCropper
            ref="cropperRef"
            :img="imgUrl"
            :info="false"
            :outputSize="1"
            outputType="png"
            :canScale="true"
            :autoCrop="true"
            :autoCropWidth="999999"
            :autoCropHeight="999999"
            :canMove="true"
            :fixedBox="false"
            :fixed="true"
            :original="false"
            :centerBox="true"
            :canMoveBox="true"
            :fixedNumber="[1, 1]"
            :limitMinSize="80"
          />
        </div>
        <br />
        <NButton type="primary" size="large" class="mx-auto block" @click="handleConfirm">
          确定
        </NButton>
      </NFlex>
    </NCard>
  </NModal>
</template>

<script setup lang="ts">
import { ref, watch } from 'vue'
import { NButton, NCard, NFlex, NModal } from 'naive-ui'
import 'vue-cropper/dist/index.css'
import { VueCropper } from 'vue-cropper'

const props = defineProps<{
  show: boolean
  image: Blob | null
}>()

const emit = defineEmits<{
  'update:show': [value: boolean]
  crop: [data: Blob]
}>()

const cropperRef = ref<any>(null)
const imgUrl = ref('')

watch(
  () => props.image,
  (img) => {
    if (img) {
      imgUrl.value = URL.createObjectURL(img)
    }
  }
)

const handleConfirm = () => {
  cropperRef.value?.getCropBlob((data: Blob) => {
    emit('crop', data)
    emit('update:show', false)
  })
}
</script>
