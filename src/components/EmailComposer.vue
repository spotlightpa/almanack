<script>
import { computed, reactive, toRefs } from "@vue/composition-api";
import { useClient } from "@/api/hooks.js";

export default {
  name: "EmailComposer",

  props: {
    initialSubject: String,
    initialBody: String,
  },
  setup(props) {
    let client = useClient();

    let rows = props.initialBody.split("\n").length;
    if (rows < 4) {
      rows = 4;
    }

    let emailStatus = reactive({
      subject: props.initialSubject,
      body: props.initialBody,
      error: null,
      isSending: false,
      hasSent: false,
    });

    return {
      ...toRefs(emailStatus),

      rows,

      hasChanged: computed(
        () =>
          props.initialSubject !== emailStatus.subject ||
          props.initialBody !== emailStatus.body
      ),
      async send() {
        if (window.confirm("Are you sure you want to send this message?")) {
          emailStatus.isSending = true;
          [, emailStatus.error] = await client.sendMessage({
            subject: emailStatus.subject,
            body: emailStatus.body,
          });
          emailStatus.isSending = false;
          if (emailStatus.error) return;
          emailStatus.hasSent = true;
          window.setTimeout(() => {
            emailStatus.hasSent = false;
          }, 5000);
        }
      },
      discard() {
        if (window.confirm("Are you sure you want to discard your changes?")) {
          emailStatus.subject = props.initialSubject;
          emailStatus.body = props.initialBody;
        }
      },
    };
  },
};
</script>
<template>
  <div class="box">
    <div class="field">
      <label class="label">Subject</label>
      <div class="control">
        <input v-model="subject" class="input" />
      </div>
      <label class="label">Body</label>
      <textarea v-model="body" class="textarea" :rows="rows"></textarea>
    </div>
    <ErrorSimple :error="error" />
    <div class="buttons">
      <button
        class="button has-text-weight-semibold is-primary"
        :class="{ 'is-loading': isSending }"
        @click="send"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'paper-plane']" />
        </span>
        <span> Send Message </span>
      </button>
      <button
        class="button has-text-weight-semibold is-danger"
        :disabled="isSending || !hasChanged"
        @click="discard"
      >
        <span class="icon">
          <font-awesome-icon :icon="['fas', 'trash-alt']" />
        </span>
        <span> Discard Changes </span>
      </button>
      <button
        class="button has-text-weight-semibold is-light"
        @click="$emit('hide')"
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
