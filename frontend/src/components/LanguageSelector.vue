<script setup lang="ts">
import { useI18n } from 'vue-i18n'
import { setLocale, SUPPORTED_LOCALES, type Locale } from '../i18n'

const { locale } = useI18n()

const labels: Record<Locale, string> = {
  en: 'EN',
  zh: '中',
  ru: 'RU',
}

function handleChange(target: Locale) {
  setLocale(target)
}
</script>

<template>
  <div class="lang-selector">
    <button
      v-for="code in SUPPORTED_LOCALES"
      :key="code"
      class="lang-btn"
      :class="{ active: locale === code }"
      @click="handleChange(code)"
    >
      {{ labels[code] }}
    </button>
  </div>
</template>

<style lang="scss" scoped>
@use '../styles/variables' as *;

.lang-selector {
  display: inline-flex;
  gap: 2px;
  background: $color-surface;
  border: 1px solid $color-border;
  border-radius: $border-radius-sm;
  padding: 2px;
}

.lang-btn {
  background: transparent;
  color: $color-text-muted;
  border: none;
  border-radius: $border-radius-sm - 1px;
  padding: 4px 10px;
  font-size: 0.8rem;
  font-weight: 600;
  cursor: pointer;
  transition: background 0.15s, color 0.15s;

  &:hover:not(.active) {
    color: $color-text;
  }

  &.active {
    background: $color-primary;
    color: $color-text;
  }
}
</style>
