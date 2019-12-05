import Vue from "vue";

import { intcomma, capfirst } from "journalize";

Vue.filter("capfirst", capfirst);
Vue.filter("intcomma", intcomma);
