import { MediaStore, Media, MediaUploadOptions, MediaList, MediaListOptions } from "tinacms";

// Declare the window.fs type
declare global {
  interface Window {
    fs: {
      writeFile: (path: string, data: ArrayBuffer) => Promise<void>;
    };
  }
}

export const mediaStore: MediaStore = {
    async persist(files: MediaUploadOptions[]) {
        const uploads = await Promise.all(
            files.map(async (fileOption) => {
                if (!fileOption.file) {
                    throw new Error('No file provided');
                }

                const file = fileOption.file;
                const filename = file.name.toLowerCase().replace(/ /g, '-');
                const path = `static/images/${filename}`;

                try {
                    // Convert file to ArrayBuffer
                    const buffer = await file.arrayBuffer();
                    await window.fs.writeFile(path, buffer);

                    // Return object matching the Media type
                    const media: Media = {
                        type: "file",
                        id: filename,
                        filename: filename,
                        directory: 'static/images',
                        src: `/images/${filename}`
                    };

                    return media;
                } catch (error) {
                    console.error('Error processing file:', error);
                    throw error;
                }
            })
        );

        return uploads;
    },

    accept: 'image/*',
    maxSize: 5 * 1024 * 1024,
    delete: function (media: Media): Promise<void> {
        throw new Error("Function not implemented.");
    },
    list: function (options?: MediaListOptions): Promise<MediaList> {
        throw new Error("Function not implemented.");
    }
};