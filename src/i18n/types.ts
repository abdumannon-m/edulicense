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
		process: string;
		partners: string;
		engagement: string;
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
		secondaryCta: string;
		scrollHint: string;
		whatsappPrefill: string;
		emailSubject: string;
		emailBody: string;
		imageAlt: string;
	};
	pillars: {
		heading: string;
		intro: string;
		items: Array<{ title: string; tagline: string; icon: PillarIcon }>;
	};
	process: {
		heading: string;
		steps: Array<{ title: string; body: string }>;
	};
	faq: {
		heading: string;
		intro: string;
		items: Array<{ question: string; answer: string }>;
	};
	partners: {
		heading: string;
		intro: string;
	};
	engagement: {
		heading: string;
		intro: string;
		items: Array<{ title: string; price: string; body: string }>;
	};
	ctaBand: {
		heading: string;
		sub: string;
		button: string;
		secondaryButton: string;
		emailSubject: string;
		emailBody: string;
		whatsappPrefill: string;
	};
	footer: {
		tagline: string;
		whatsapp: string;
		email: string;
		telegram: string;
		privacy: string;
		rights: string;
		addressLabel: string;
		address: string;
		directions: string;
		mapTitle: string;
		creditPhoto: string;
	};
	privacy: {
		title: string;
		body: string;
		back: string;
	};
}
