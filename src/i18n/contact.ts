/** Digits only, country code included (no +), for wa.me — replace with your number */
export const WHATSAPP_PHONE = '998901234567';

/** Telegram @username without @; empty string hides Telegram links */
export const TELEGRAM_USERNAME = '';

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
