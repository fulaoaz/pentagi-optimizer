import { zodResolver } from '@hookform/resolvers/zod';
import { Eye, EyeOff } from 'lucide-react';
import { useMemo, useState } from 'react';
import { useForm } from 'react-hook-form';
import { toast } from 'sonner';
import * as z from 'zod';

import type { Translate } from '@/lib/i18n';

import { Button } from '@/components/ui/button';
import { Form, FormControl, FormDescription, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';
import { FormSubmitButton } from '@/components/ui/form-submit-button';
import { Input } from '@/components/ui/input';
import { useLocale } from '@/hooks/use-locale';
import { api, type ApiErrorResponse, type ApiHttpError } from '@/lib/axios';

const buildPasswordChangeSchema = (t: Translate) =>
    z
        .object({
            confirmPassword: z.string().min(1, { message: t('auth.confirmPasswordRequired') }),
            currentPassword: z.string().min(1, { message: t('auth.currentPasswordRequired') }),
            newPassword: z
                .string()
                .min(8, { message: t('auth.passwordMinLength') })
                .max(100, { message: t('auth.passwordMaxLength') })
                .refine(
                    (password) => {
                        if (password.length > 15) {
                            return true;
                        }

                        return (
                            password.length >= 8 &&
                            /[0-9]/.test(password) &&
                            /[a-z]/.test(password) &&
                            /[A-Z]/.test(password) &&
                            /[!@#$&*]/.test(password)
                        );
                    },
                    { message: t('auth.passwordComplexity') },
                ),
        })
        .refine((data) => data.newPassword === data.confirmPassword, {
            message: t('auth.passwordsMismatch'),
            path: ['confirmPassword'],
        })
        .refine((data) => data.currentPassword !== data.newPassword, {
            message: t('auth.newPasswordDifferent'),
            path: ['newPassword'],
        });

interface PasswordChangeFormProps {
    isModal?: boolean;
    onCancel?: () => void;
    onSkip?: () => void;
    onSuccess?: () => void;
    showSkip?: boolean;
}

type PasswordChangeFormValues = z.infer<ReturnType<typeof buildPasswordChangeSchema>>;

export function PasswordChangeForm({
    isModal = true,
    onCancel,
    onSkip,
    onSuccess,
    showSkip = false,
}: PasswordChangeFormProps) {
    const { t } = useLocale();
    const [error, setError] = useState<null | string>(null);
    const [showCurrentPassword, setShowCurrentPassword] = useState(false);
    const [showNewPassword, setShowNewPassword] = useState(false);
    const [showConfirmPassword, setShowConfirmPassword] = useState(false);
    const passwordChangeSchema = useMemo(() => buildPasswordChangeSchema(t), [t]);

    const form = useForm<PasswordChangeFormValues>({
        defaultValues: {
            confirmPassword: '',
            currentPassword: '',
            newPassword: '',
        },
        resolver: zodResolver(passwordChangeSchema),
    });

    const handleSubmit = async (values: PasswordChangeFormValues) => {
        setError(null);

        try {
            await api.put('/user/password', {
                confirm_password: values.confirmPassword,
                current_password: values.currentPassword,
                password: values.newPassword,
            });

            form.reset();
            setShowCurrentPassword(false);
            setShowNewPassword(false);
            setShowConfirmPassword(false);

            toast.success(t('auth.passwordChanged'));

            if (onSuccess) {
                onSuccess();
            }
        } catch (err: unknown) {
            const error = err as ApiHttpError;
            const responseData = error.response?.data as ApiErrorResponse | undefined;

            let errorMessage = t('auth.changePasswordFailed');

            if (responseData?.msg) {
                errorMessage = responseData.msg;
            } else if (responseData?.code) {
                switch (responseData.code) {
                    case 'AuthRequired':
                        errorMessage = t('auth.authenticationRequired');
                        break;
                    case 'Users.ChangePasswordCurrentUser.InvalidCurrentPassword':
                        errorMessage = t('auth.currentPasswordIncorrect');
                        break;
                    case 'Users.ChangePasswordCurrentUser.InvalidNewPassword':
                        errorMessage = t('auth.newPasswordInvalid');
                        break;
                    case 'Users.ChangePasswordCurrentUser.InvalidPassword':
                        errorMessage = t('auth.passwordValidationFailed');
                        break;
                    case 'Users.NotFound':
                        errorMessage = t('auth.userNotFound');
                        break;
                    default:
                        errorMessage = responseData.msg || error.message || t('auth.changePasswordFailed');
                }
            } else if (error.message) {
                errorMessage = error.message;
            }

            setError(errorMessage);
        }
    };

    return (
        <Form {...form}>
            <form
                className="flex flex-col gap-4"
                onSubmit={form.handleSubmit(handleSubmit)}
            >
                <FormField
                    control={form.control}
                    name="currentPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>{t('auth.currentPassword')}</FormLabel>
                            <FormControl>
                                <div className="relative">
                                    <Input
                                        {...field}
                                        placeholder={t('auth.enterCurrentPassword')}
                                        type={showCurrentPassword ? 'text' : 'password'}
                                    />
                                    <Button
                                        aria-label={
                                            showCurrentPassword ? t('auth.hidePassword') : t('auth.showPassword')
                                        }
                                        className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() => setShowCurrentPassword(!showCurrentPassword)}
                                        size="sm"
                                        tabIndex={-1}
                                        type="button"
                                        variant="ghost"
                                    >
                                        {showCurrentPassword ? (
                                            <EyeOff className="text-muted-foreground size-4" />
                                        ) : (
                                            <Eye className="text-muted-foreground size-4" />
                                        )}
                                    </Button>
                                </div>
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <FormField
                    control={form.control}
                    name="newPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>{t('auth.newPassword')}</FormLabel>
                            <FormControl>
                                <div className="relative">
                                    <Input
                                        {...field}
                                        placeholder={t('auth.enterNewPassword')}
                                        type={showNewPassword ? 'text' : 'password'}
                                    />
                                    <Button
                                        aria-label={showNewPassword ? t('auth.hidePassword') : t('auth.showPassword')}
                                        className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() => setShowNewPassword(!showNewPassword)}
                                        size="sm"
                                        tabIndex={-1}
                                        type="button"
                                        variant="ghost"
                                    >
                                        {showNewPassword ? (
                                            <EyeOff className="text-muted-foreground size-4" />
                                        ) : (
                                            <Eye className="text-muted-foreground size-4" />
                                        )}
                                    </Button>
                                </div>
                            </FormControl>
                            <FormDescription className="text-xs">{t('auth.passwordComplexityHint')}</FormDescription>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                <FormField
                    control={form.control}
                    name="confirmPassword"
                    render={({ field }) => (
                        <FormItem>
                            <FormLabel>{t('auth.confirmNewPassword')}</FormLabel>
                            <FormControl>
                                <div className="relative">
                                    <Input
                                        {...field}
                                        placeholder={t('auth.confirmNewPasswordPlaceholder')}
                                        type={showConfirmPassword ? 'text' : 'password'}
                                    />
                                    <Button
                                        aria-label={
                                            showConfirmPassword ? t('auth.hidePassword') : t('auth.showPassword')
                                        }
                                        className="absolute top-0 right-0 h-full px-3 py-2 hover:bg-transparent"
                                        onClick={() => setShowConfirmPassword(!showConfirmPassword)}
                                        size="sm"
                                        tabIndex={-1}
                                        type="button"
                                        variant="ghost"
                                    >
                                        {showConfirmPassword ? (
                                            <EyeOff className="text-muted-foreground size-4" />
                                        ) : (
                                            <Eye className="text-muted-foreground size-4" />
                                        )}
                                    </Button>
                                </div>
                            </FormControl>
                            <FormMessage />
                        </FormItem>
                    )}
                />

                {error && <div className="text-destructive text-sm">{error}</div>}

                <div className="flex justify-end gap-2 pt-2">
                    {showSkip && (
                        <Button
                            className="text-muted-foreground"
                            onClick={onSkip}
                            type="button"
                            variant="ghost"
                        >
                            {t('auth.skipForNow')}
                        </Button>
                    )}
                    {isModal && (
                        <Button
                            onClick={onCancel}
                            type="button"
                            variant="outline"
                        >
                            {t('common.cancel')}
                        </Button>
                    )}
                    <FormSubmitButton>
                        <span>{t('auth.updatePassword')}</span>
                    </FormSubmitButton>
                </div>
            </form>
        </Form>
    );
}
