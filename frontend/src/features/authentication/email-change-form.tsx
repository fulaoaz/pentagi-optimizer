
import { useMemo, useState } from 'react';
import { toast } from 'sonner';
import * as z from 'zod';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { FormSubmitButton } from '@/components/ui/form-submit-button';
import { Input } from '@/components/ui/input';
import { InputPassword } from '@/components/ui/input-password';
import { useAppForm } from '@/hooks/use-app-form';
import { useLocale } from '@/hooks/use-locale';
import type { Translate } from '@/lib/i18n';
import { api, resolveApiErrorMessage } from '@/lib/axios';
import { useUser } from '@/providers/user-provider';

const buildEmailChangeSchema = (t: Translate) =>
    z.object({
        currentPassword: z.string().min(1, { message: t('auth.currentPasswordRequired') }),
        newEmail: z
            .string()
            .trim()
            .toLowerCase()
            .min(1, { message: t('auth.emailRequired') })
            .email({ message: t('auth.invalidEmail') })
            .max(50, { message: t('auth.emailMaxLength') }),
    });

const ERROR_BY_CODE: Record<string, string> = {
    'Users.ChangeEmailCurrentUser.EmailAlreadyExists': 'Email address is already in use',
    'Users.ChangeEmailCurrentUser.InvalidCurrentPassword': 'Current password is incorrect',
    'Users.ChangeEmailCurrentUser.InvalidEmail': 'New email does not meet requirements',
    'Users.NotFound': 'User not found',
};

interface EmailChangeFormProps {
    onCancel?: () => void;
    onSuccess?: () => void;
}

type EmailChangeFormValues = z.infer<ReturnType<typeof buildEmailChangeSchema>>;

export function EmailChangeForm({ onCancel, onSuccess }: EmailChangeFormProps) {
    const { t } = useLocale();
    const [error, setError] = useState<null | string>(null);
    const { patchUser, refreshAuthInfo } = useUser();
    const emailChangeSchema = useMemo(() => buildEmailChangeSchema(t), [t]);

    const form = useAppForm<EmailChangeFormValues>({
        defaultValues: {
            currentPassword: '',
            newEmail: '',
        },
        schema: emailChangeSchema,
    });

    const handleSubmit = async (values: EmailChangeFormValues) => {
        setError(null);

        try {
            await api.put('/user/email', {
                current_password: values.currentPassword,
                mail: values.newEmail,
            });

            form.reset();
            toast.success(t('auth.emailUpdated'));

            patchUser({ mail: values.newEmail });
            await refreshAuthInfo();

            onSuccess?.();
        } catch (err: unknown) {
            setError(resolveApiErrorMessage(err, ERROR_BY_CODE, 'Failed to update email'));
        }
    };

    return (
        <Form {...form}>
            {/* noValidate: the type="email" field would otherwise fire the browser's native (locale-styled)
                validation popup on submit, pre-empting our zod message. Validation runs through zod instead. */}
            <form
                className="flex flex-col gap-4"
                noValidate
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <FormField
                    control={form.control}
                    name="currentPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>{t('auth.currentPassword')}</FormLabel>
                            <FormControl>
                                <InputPassword
                                    {...field}
                                    placeholder={t('auth.enterCurrentPassword')}
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <FormField
                    control={form.control}
                    name="newEmail"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>{t('auth.newEmail')}</FormLabel>
                            <FormControl>
                                <Input
                                    {...field}
                                    placeholder={t('auth.enterNewEmail')}
                                    type="email"
                                />
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                {error && <div className="text-destructive text-sm">{error}</div>}

                <div className="flex justify-end gap-2 pt-2">
                    {onCancel && (
                        <Button
                            onClick={onCancel}
                            size="sm"
                            type="button"
                            variant="outline"
                        >
                            {t('common.cancel')}
                        </Button>
                    )}
                    <FormSubmitButton size="sm">
                        <span>{t('auth.updateEmail')}</span>
                    </FormSubmitButton>
                </div>
            </form>
        </Form>
    );
}
