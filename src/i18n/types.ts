export type Locale = 'en' | 'uz';

export type PillarIcon = 'license' | 'institution' | 'operations';

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
		menu: string;
		menuOpen: string;
		menuClose: string;
	};
	hero: {
		eyebrow: string;
		title: string;
		subtitle: string;
		cta: string;
		scrollHint: string;
		whatsappPrefill: string;
		imageAlt: string;
	};
	pillars: {
		heading: string;
		intro: string;
		items: Array<{ title: string; tagline: string; icon: PillarIcon }>;
	};
	audience: {
		heading: string;
		items: string[];
	};
	why: {
		heading: string;
		body: string;
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
		creditPhoto: string;
	};
	privacy: {
		title: string;
		body: string;
		back: string;
	};
}
