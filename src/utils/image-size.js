export default function imageSize(url) {
  return new Promise((resolve, reject) => {
    let img = new Image();
    img.onload = () => {
      resolve({
        height: img.height,
        width: img.width,
      });
    };
    img.onerror = (e) => reject(e);
    img.src = url;
  });
}
