<script>
import "@/css/main.scss";

import PageNav from "./PageNav.vue";

export default {
  components: {
    PageNav,
  },
};
</script>

<template>
  <div class="body">
    <PageNav id="top-nav" />
    <main class="section container">
      <div class="columns is-centered is-multiline">
        <div class="column is-four-fifths">
          <router-view />
        </div>
      </div>
    </main>
    <section class="footer is-size-4">
      <div class="columns is-centered is-multiline">
        <div class="column is-two-fifths has-text-centered is-size-5">
          <p>
            © <span v-text="new Date().getFullYear()"></span> <br />
            Spotlight PA <br />
            225 Market St Suite 502A <br />
            Harrisburg PA 17101
          </p>
          <p>
            <em>
              All Spotlight PA content is provided “as-is” and its use is at the
              sole risk and discretion of partner news organizations. To view
              the full licensing agreement,
              <a href="/docs/Spotlight PA Content License.docx" download
                >click here</a
              >.
            </em>
          </p>
        </div>
      </div>
    </section>
  </div>
</template>
