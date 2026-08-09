/** Digits only, country code included (no +), for wa.me — replace with your number */
export const WHATSAPP_PHONE = '998901234567';

/** Telegram @username without @; empty string hides Telegram links */
export const TELEGRAM_USERNAME = '';

/** Public inbox for structured inquiries. */
export const CONTACT_EMAIL = 'info@edulicense.uz';

/** Office coordinates (Sayram street, 7th lane 21) used by the Yandex Maps embed. */
export const OFFICE_LAT = 41.325973;
export const OFFICE_LON = 69.317971;

/** Shareable Yandex Maps link to the office pin. */
export const OFFICE_MAP_URL = 'https://yandex.uz/maps/-/CTS0uF-f';

/**
 * Keyless Yandex Maps widget embed for the office pin.
 * `locale` is a Yandex language tag, e.g. `en_US` or `uz_UZ`.
 */
export function officeMapEmbedUrl(locale: string, zoom = 17): string {
	const params = new URLSearchParams({
		ll: `${OFFICE_LON},${OFFICE_LAT}`,
		z: String(zoom),
		pt: `${OFFICE_LON},${OFFICE_LAT},pm2rdm`,
		lang: locale,
	});
	return `https://yandex.uz/map-widget/v1/?${params.toString()}`;
}

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
