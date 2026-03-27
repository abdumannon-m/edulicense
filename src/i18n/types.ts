export type Locale = 'en' | 'uz';

export interface Messages {
	lang: Locale;
	ogLocale: string;
	metaTitle: string;
	metaDescription: string;
	brandShort: string;
	brandLegal: string;
	nav: {
		services: string;
		audience: string;
		why: string;
		process: string;
		faq: string;
		contact: string;
		language: string;
		switchToEn: string;
		switchToUz: string;
	};
	hero: {
		title: string;
		subtitle: string;
		cta: string;
		scrollHint: string;
		whatsappPrefill: string;
	};
	pillars: {
		heading: string;
		intro: string;
		items: Array<{ title: string; bullets: string[] }>;
	};
	audience: {
		heading: string;
		items: string[];
	};
	why: {
		heading: string;
		paragraphs: string[];
	};
	process: {
		heading: string;
		steps: Array<{ title: string; body: string }>;
	};
	faq: {
		heading: string;
		items: Array<{ question: string; answer: string }>;
	};
	ctaBand: {
		heading: string;
		sub: string;
		button: string;
		whatsappPrefill: string;
	};
	footer: {
		tagline: string;
		whatsapp: string;
		telegram: string;
		privacy: string;
		rights: string;
		addressLine: string;
	};
	privacy: {
		title: string;
		body: string;
		back: string;
	};
}
