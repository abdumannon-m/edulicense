/** Digits only, country code included (no +), for wa.me — replace with your number */
export const WHATSAPP_PHONE = '998901234567';

/** Telegram @username without @; empty string hides Telegram links */
export const TELEGRAM_USERNAME = '';

/** Public inbox for structured inquiries. */
export const CONTACT_EMAIL = 'info@edulicense.uz';

export function whatsappUrl(prefillMessage: string): string {
	const phone = WHATSAPP_PHONE.replace(/\D/g, '');
	return `https://wa.me/${phone}?text=${encodeURIComponent(prefillMessage)}`;
}

export function telegramUrl(): string {
	const u = TELEGRAM_USERNAME.replace(/^@/, '');
	return `https://t.me/${u}`;
}

export function telegramEnabled(): boolean {
	return TELEGRAM_USERNAME.trim().length > 0;
}

export function emailUrl(subject: string, body = ''): string {
	const params = new URLSearchParams({ subject });
	if (body) {
		params.set('body', body);
	}
	return `mailto:${CONTACT_EMAIL}?${params.toString()}`;
}
