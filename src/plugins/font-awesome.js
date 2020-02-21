import Vue from "vue";
import { library } from "@fortawesome/fontawesome-svg-core";
import { FontAwesomeIcon } from "@fortawesome/vue-fontawesome";
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
  faPenNib,
  faSyncAlt,
  faUserCircle,
  faUserClock,
} from "@fortawesome/free-solid-svg-icons";

// Buefy icons
import {
  faAngleDown,
  faAngleLeft,
  faAngleRight,
  faArrowUp,
  faCaretDown,
  faCaretUp,
  faCheck,
  faCheckCircle,
  faExclamationCircle,
  faExclamationTriangle,
  faEye,
  faEyeSlash,
  faInfoCircle,
  faUpload,
} from "@fortawesome/free-solid-svg-icons";

library.add(
  faAngleDown,
  faAngleLeft,
  faAngleRight,
  faArrowUp,
  faCaretDown,
  faCaretUp,
  faCheck,
  faCheckCircle,
  faCopy,
  faExclamationCircle,
  faExclamationTriangle,
  faEye,
  faEyeSlash,
  faFileCode,
  faFileDownload,
  faFileUpload,
  faFileWord,
  faInfoCircle,
  faLink,
  faNewspaper,
  faPenNib,
  faSyncAlt,
  faUpload,
  faUserCircle,
  faUserClock
);

Vue.component("font-awesome-icon", FontAwesomeIcon);
