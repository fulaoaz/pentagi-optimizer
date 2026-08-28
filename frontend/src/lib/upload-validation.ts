/**
 * Shared client-side preflight for multipart uploads. Run before constructing
 * the FormData so the user gets an instant toast instead of waiting for the
 * network round-trip and a generic 4xx error.
 *
 * Mirrors the per-file / per-batch limits enforced by the backend
 * (`pkg/resources/resources.go` for the resources library and
 * `pkg/flowfiles/files.go` for the flow files cache). Both backends do not
 * whitelist file extensions, so neither does this validator — it only checks
 * size, count and emptiness.
 */

export interface UploadValidationLimits {
    /** Maximum number of files allowed per request. */
    maxFiles: number;
    /** Maximum size of a single file in megabytes. */
    maxFileSizeMb: number;
    /** Maximum combined size of the batch in megabytes. */
    maxTotalSizeMb: number;
    /**
     * Reject 0-byte files. Defaults to `true` because both servers stream the
     * upload body and surface a confusing `EOF on read` error mid-request when
     * an empty file lands in the multipart payload.
     */
    rejectEmpty?: boolean;
}

export interface UploadValidationMessages {
    emptyFile: (name: string) => string;
    fileTooLarge: (name: string, maxFileSizeMb: number) => string;
    tooManyFiles: (maxFiles: number) => string;
    totalTooLarge: (maxTotalSizeMb: number) => string;
}

const MEGABYTE = 1024 * 1024;

/**
 * Validate a batch of `File` objects against the supplied limits. Returns
 * `null` when the batch is acceptable or the user-facing error message for
 * the **first** violation otherwise — callers typically forward this string
 * directly into a `toast.error('Upload failed', { description })`.
 *
 * Empty batches are treated as a no-op (`null`); callers usually short-circuit
 * before reaching the validator anyway.
 */
export const validateUploadBatch = (
    files: readonly File[],
    limits: UploadValidationLimits,
    messages?: UploadValidationMessages,
): null | string => {
    if (files.length > limits.maxFiles) {
        return messages?.tooManyFiles(limits.maxFiles) ?? `Too many files: max ${limits.maxFiles} per upload`;
    }

    const maxBytesPerFile = limits.maxFileSizeMb * MEGABYTE;
    const maxTotalBytes = limits.maxTotalSizeMb * MEGABYTE;
    const rejectEmpty = limits.rejectEmpty ?? true;
    let totalBytes = 0;

    for (const file of files) {
        if (rejectEmpty && file.size === 0) {
            return messages?.emptyFile(file.name) ?? `File "${file.name}" is empty`;
        }

        if (file.size > maxBytesPerFile) {
            return (
                messages?.fileTooLarge(file.name, limits.maxFileSizeMb) ??
                `File "${file.name}" is larger than ${limits.maxFileSizeMb} MB`
            );
        }

        totalBytes += file.size;
    }

    if (totalBytes > maxTotalBytes) {
        return (
            messages?.totalTooLarge(limits.maxTotalSizeMb) ??
            `Total upload size exceeds the ${limits.maxTotalSizeMb} MB limit`
        );
    }

    return null;
};
