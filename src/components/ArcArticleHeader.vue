<script>
import { h } from "vue";

function ArcArticleHeader(props) {
  return h(`h${props.block.level}`, { innerHTML: props.block.content });
}

ArcArticleHeader.props = {
  block: Object,
};

export default ArcArticleHeader;
</script>
