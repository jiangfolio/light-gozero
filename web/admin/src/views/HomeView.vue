<script setup lang="ts">
import { useAppStore } from '@/stores/app'

const app = useAppStore()
</script>

<template>
  <div class="mx-auto flex min-h-screen max-w-3xl flex-col gap-6 p-8">
    <div class="flex items-center gap-3">
      <IconParkHome class="text-[28px] text-[#1677ff]" />
      <ATypographyTitle :level="3" class="!mb-0">{{ app.title }}</ATypographyTitle>
    </div>
    <ATypographyParagraph type="secondary">
      Vite + Vue3 + TS + Tailwind + Pinia + Axios + antdv-next + IconPark（自动导入 / 离线）
    </ATypographyParagraph>
    <ASpace>
      <AButton type="primary">antdv-next Button</AButton>
      <AButton>
        <IconParkSettingTwo class="mr-1 align-middle text-[16px]" />
        IconPark
      </AButton>
    </ASpace>
  </div>
</template>
