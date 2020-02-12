import Vue from "vue";
import { library } from "@fortawesome/fontawesome-svg-core";
import {
  faCopy,
  faFileCode,
  faFileWord,
  faNewspaper,
} from "@fortawesome/free-regular-svg-icons";
import {
  faFileDownload,
  faLink,
  faFileUpload,
  faSyncAlt,
  faUserClock,
} from "@fortawesome/free-solid-svg-icons";

// Buefy internal icons
import {
  faCalendarAlt,
  faCheck,
  faCheckCircle,
  faInfoCircle,
  faExclamationTriangle,
  faExclamationCircle,
  faArrowUp,
  faAngleRight,
  faAngleLeft,
  faAngleDown,
  faEye,
  faEyeSlash,
  faCaretDown,
  faCaretUp,
  faUpload,
} from "@fortawesome/free-solid-svg-icons";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";

library.add(
  faCalendarAlt,
  faCheck,
  faCheckCircle,
  faInfoCircle,
  faExclamationTriangle,
  faExclamationCircle,
  faArrowUp,
  faAngleRight,
  faAngleLeft,
  faAngleDown,
  faEye,
  faEyeSlash,
  faCaretDown,
  faCaretUp,
  faUpload,
  faCopy,
  faFileCode,
  faFileWord,
  faNewspaper,
  faFileDownload,
  faLink,
  faFileUpload,
  faSyncAlt,
  faUserClock
);

Vue.component("font-awesome-icon", FontAwesomeIcon);
