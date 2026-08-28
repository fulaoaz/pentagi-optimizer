import { Replace } from 'lucide-react';

import type { Translate } from '@/lib/i18n';

import ConfirmationDialog from '@/components/shared/confirmation-dialog';
import { useLocale } from '@/hooks/use-locale';

export interface OverwriteConflict {
    destination: string;
    /** Display name extracted from `destination` for the confirm dialog. */
    destinationName: string;
}

interface OverwriteDialogProps {
    /**
     * Overrides the auto-generated confirm button label. Defaults to
     * `"Replace"` for a single conflict and `"Replace all"` for a batch.
     */
    confirmText?: string;
    /**
     * Conflicts collected from a batch operation. Empty array keeps the dialog hidden.
     * For a single conflict the message names the conflicting item; for many it falls
     * back to a count-based summary (Finder-style "Apply to all").
     */
    conflicts: OverwriteConflict[];
    /**
     * Optional override for the description body. When omitted, the component
     * renders the canonical Finder-style copy that names the single conflicting
     * item or the count for a batch.
     */
    description?: string;
    onCancel: () => void;
    onReplaceAll: () => Promise<unknown> | unknown;
    /** Optional override for the dialog title. Defaults to `"Replace existing item?"`. */
    title?: string;
}

const buildDefaultDescription = (conflicts: OverwriteConflict[], t: Translate): string | undefined => {
    const single = conflicts.length === 1 ? conflicts[0] : undefined;

    if (single) {
        return t('overwrite.descriptionOne', {
            destination: single.destination,
            name: single.destinationName,
        });
    }

    if (conflicts.length > 1) {
        return t('overwrite.descriptionMany', { count: conflicts.length });
    }

    return undefined;
};

const buildDefaultConfirmText = (count: number, t: Translate): string =>
    count > 1 ? t('overwrite.replaceAll') : t('overwrite.replace');

/**
 * Shared "Replace or cancel" confirmation for destructive overwrite flows
 * (move / copy / pull / attach / promote, …). The hook owns the conflict
 * state; this component only renders the prompt and forwards the user's
 * decision back through the callbacks. A batch decision (Replace all) is
 * applied to every pending conflict in one shot — this matches the OS
 * file-manager UX and keeps the user from being prompted N times for the
 * same destination directory.
 */
export function OverwriteDialog({
    confirmText,
    conflicts,
    description,
    onCancel,
    onReplaceAll,
    title,
}: OverwriteDialogProps) {
    const { t } = useLocale();

    return (
        <ConfirmationDialog
            cancelText={t('common.cancel')}
            confirmIcon={<Replace />}
            confirmText={confirmText ?? buildDefaultConfirmText(conflicts.length, t)}
            confirmVariant="destructive"
            description={description ?? buildDefaultDescription(conflicts, t)}
            handleConfirm={async () => {
                await onReplaceAll();
            }}
            handleOpenChange={(nextOpen) => {
                if (!nextOpen) {
                    onCancel();
                }
            }}
            isOpen={conflicts.length > 0}
            title={title ?? t('overwrite.title')}
        />
    );
}
