import JSZip from "jszip";
import { FileForArchive } from "../types/types";
import crypto from "crypto";

export default async function archiveFiles(files: FileForArchive[], userEmail: string): Promise<FileForArchive | undefined> {
    try {
        const zip = new JSZip();
    
        await Promise.all(files.map(file => {
            zip.file(file.fileName, file.buffer, {binary: true})
        }));
    
        return zip.generateAsync({type: "nodebuffer"}).then(data => {
            return {
                fileName: `archive_${crypto.randomUUID()}.zip`,
                buffer: data
            }
        });
    } catch (err: any) {
        throw new Error(err);
    };
};