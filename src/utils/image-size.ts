export interface Dimensions {
  width: number;
  height: number;
}

/**
 * Measure an image from a URL.
 * Rejects if the image cannot be loaded.
 */
export default function imageSize(url: string): Promise<Dimensions> {
  return new Promise((resolve, reject) => {
    let img = new Image();
    img.onload = () => {
      resolve({
        height: img.naturalHeight,
        width: img.naturalWidth,
      });
    };
    img.onerror = (e) => reject(e);
    img.src = url;
  });
}

/**
 * Measure an image from a File object (before upload).
 * Returns {width: 0, height: 0} if the image cannot be decoded.
 */
export async function imageFileSize(file: File): Promise<Dimensions> {
  const url = URL.createObjectURL(file);
  try {
    return await imageSize(url);
  } catch {
    return { width: 0, height: 0 };
  } finally {
    URL.revokeObjectURL(url);
  }
}
