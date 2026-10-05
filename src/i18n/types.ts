export type Locale = 'en' | 'uz';

/** Status of a line on one of the example cards. */
export type CardState = 'done' | 'active' | 'next' | 'gap';

export interface CardRow {
	label: string;
	value: string;
	state: CardState;
}

export interface Messages {
	lang: Locale;
	ogLocale: string;
	metaTitle: string;
	metaDescription: string;
	brandShort: string;
	brandLegal: string;
	nav: {
		how: string;
		services: string;
		partners: string;
		pricing: string;
		faq: string;
		contact: string;
		switchToEn: string;
		switchToUz: string;
		menuOpen: string;
		menuClose: string;
		sections: string;
	};
	hero: {
		title: string;
		subtitle: string;
		cta: string;
		secondaryCta: string;
		emailSubject: string;
		emailBody: string;
		whatsappPrefill: string;
	};
	/** The example licensing roadmap shown under the hero. */
	roadmap: {
		title: string;
		example: string;
		school: string;
		progress: string;
		steps: CardRow[];
		nextLabel: string;
		next: string;
	};
	how: {
		heading: string;
		intro: string;
		items: Array<{
			step: string;
			title: string;
			body: string;
			cardTitle: string;
			rows: CardRow[];
		}>;
	};
	services: {
		heading: string;
		intro: string;
		optionTag: string;
		items: Array<{ title: string; body: string; chips: string[]; option?: boolean }>;
	};
	partners: {
		heading: string;
		newTab: string;
	};
	stats: {
		heading: string;
		items: Array<{ value: string; label: string; body: string }>;
	};
	engagement: {
		heading: string;
		intro: string;
		items: Array<{ title: string; price: string; body: string }>;
	};
	faq: {
		heading: string;
		items: Array<{ question: string; answer: string }>;
	};
	cta: {
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
		contactLabel: string;
		pagesLabel: string;
	};
	privacy: {
		title: string;
		body: string;
		back: string;
	};
	stateLabels: Record<CardState, string>;
}
