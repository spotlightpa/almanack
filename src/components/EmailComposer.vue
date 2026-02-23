<script setup>
import { computed, ref } from "vue";
import { post, sendMessage } from "@/api/client-v2.js";

const props = defineProps({
  subject: { type: String, required: true },
  body: { type: String, required: true },
  canDiscard: { type: Boolean, default: false },
});

const emit = defineEmits(["update:subject", "update:body", "hide", "discard"]);

const rows = computed(() => Math.max(4, props.body.split("\n").length));

const error = ref(null);
const isSending = ref(false);
const hasSent = ref(false);

async function send() {
  if (window.confirm("Are you sure you want to send this message?")) {
    isSending.value = true;
    [, error.value] = await post(sendMessage, {
      subject: props.subject,
      body: props.body,
    });
    isSending.value = false;
    if (error.value) return;
    hasSent.value = true;
    window.setTimeout(() => {
      hasSent.value = false;
    }, 5000);
  }
}
</script>

<template>
  <div class="box">
    <div class="field">
      <BulmaFieldInput
        :model-value="subject"
        label="Subject"
        @update:model-value="emit('update:subject', $event)"
      ></BulmaFieldInput>
      <BulmaTextarea
        :model-value="body"
        label="Body"
        :rows="rows"
        @update:model-value="emit('update:body', $event)"
      ></BulmaTextarea>
    </div>
    <ErrorSimple :error="error"></ErrorSimple>
    <div class="buttons">
      <button
        class="button has-text-weight-semibold is-primary"
        :class="{ 'is-loading': isSending }"
        @click="send"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'paper-plane']"></font-awesome-icon>
        </span>
        <span> Send Message </span>
      </button>
      <button
        v-if="canDiscard"
        class="button has-text-weight-semibold is-danger"
        :disabled="isSending || null"
        @click="emit('discard')"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'trash-alt']"></font-awesome-icon>
        </span>
        <span> Discard Changes </span>
      </button>
      <button
        class="button has-text-weight-semibold is-light"
        @click="emit('hide')"
      >
        Close Composer
      </button>
      <transition name="fade">
        <div
          v-if="hasSent"
          class="tag is-rounded is-success is-light has-text-weight-semibold"
        >
          Message Sent
        </div>
      </transition>
    </div>
  </div>
</template>

<style scoped>
.fade {
  transition: all 0.5s ease;
}
</style>
