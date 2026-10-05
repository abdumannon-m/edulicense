import type { Messages } from './types';

export const en: Messages = {
	lang: 'en',
	ogLocale: 'en_GB',
	metaTitle: 'Edu License · International school licensing in Uzbekistan',
	metaDescription:
		'We take schools and universities in Uzbekistan from a first readiness review to international authorisation: licensing, institutional setup and launch.',
	brandShort: 'Edu License',
	brandLegal: 'Edu License LLC',
	nav: {
		how: 'How it works',
		services: 'Services',
		partners: 'Partners',
		pricing: 'Pricing',
		faq: 'FAQ',
		contact: 'Email us',
		switchToEn: 'English',
		switchToUz: "O'zbekcha",
		menuOpen: 'Open menu',
		menuClose: 'Close menu',
		sections: 'Page sections',
	},
	hero: {
		title: 'Get your school internationally licensed',
		subtitle:
			'A Tashkent team that takes schools and universities from a first readiness review to authorisation, and stays until you open.',
		cta: 'Email the brief',
		secondaryCta: 'Message on WhatsApp',
		emailSubject: 'Institution readiness assessment',
		emailBody:
			'Hello Edu License,\n\nWe would like to discuss international licensing for our institution.\n\nInstitution name:\nCity:\nCurrent licence status:\nTarget programme:\nDesired timeline:\nMain contact:\n',
		whatsappPrefill:
			'Hello, I would like to discuss international licensing for our institution in Uzbekistan.',
	},
	roadmap: {
		title: 'Licensing roadmap',
		example: 'Example',
		school: 'Your school · Tashkent',
		progress: '2 of 5 steps done',
		steps: [
			{ label: 'Readiness review', value: 'Gaps found: 6', state: 'done' },
			{ label: 'Programme fit', value: 'Cambridge pathway', state: 'done' },
			{ label: 'Evidence and documents', value: '12 of 20 ready', state: 'active' },
			{ label: 'Submission', value: 'Planned for March', state: 'next' },
			{ label: 'Authorisation', value: 'Board decision', state: 'next' },
		],
		nextLabel: 'Next up',
		next: 'Governance handbook draft, reviewed with your principal on Friday.',
	},
	how: {
		heading: 'How it works',
		intro: 'Three steps, one team beside you the whole way.',
		items: [
			{
				step: '01',
				title: 'We review where you stand',
				body: 'One session on your licence, website, policies and people. You leave with a list of gaps and an honest view of what fits.',
				cardTitle: 'Readiness review',
				rows: [
					{ label: 'School licence', value: 'In order', state: 'done' },
					{ label: 'Curriculum documents', value: '3 gaps', state: 'gap' },
					{ label: 'Safeguarding policy', value: 'Missing', state: 'gap' },
					{ label: 'Teacher qualifications', value: 'In order', state: 'done' },
				],
			},
			{
				step: '02',
				title: 'We map the shortest path',
				body: 'A sequenced plan your team can follow: who does what, which documents, by when.',
				cardTitle: 'Roadmap',
				rows: [
					{ label: 'October', value: 'Governance and handbooks', state: 'active' },
					{ label: 'December', value: 'Assessment and admissions', state: 'next' },
					{ label: 'February', value: 'Mock inspection', state: 'next' },
					{ label: 'March', value: 'Submission', state: 'next' },
				],
			},
			{
				step: '03',
				title: 'We build it with your team',
				body: 'We draft, review and prepare alongside your leadership until the evidence is real, then hand over playbooks for launch.',
				cardTitle: 'Evidence pack',
				rows: [
					{ label: 'Governance handbook', value: 'Approved', state: 'done' },
					{ label: 'Assessment policy', value: 'In review', state: 'active' },
					{ label: 'Staff handbook', value: 'Drafting', state: 'active' },
					{ label: 'Admissions policy', value: 'Next', state: 'next' },
				],
			},
		],
	},
	services: {
		heading: 'What we cover',
		intro: 'Pick one area or the whole path. We confirm fit in the first conversation.',
		optionTag: 'Option',
		items: [
			{
				title: 'International licensing',
				body: 'Map your path to authorisation with the framework that fits your school, with fewer gaps and clearer evidence.',
				chips: ['Cambridge', 'IB-oriented', 'American', 'Dual diploma'],
			},
			{
				title: 'Institutional setup',
				body: 'Governance, handbooks and academic systems that match what inspectors and partners expect to see.',
				chips: ['Governance', 'Handbooks', 'Academic systems'],
			},
			{
				title: 'Operational readiness',
				body: 'Roles, timetables and launch rhythms, so your team is ready for day one and not only on paper.',
				chips: ['Roles', 'Timetables', 'Launch plan'],
			},
			{
				title: 'SAT test centre',
				body: 'Become an official SAT test centre: we handle the CEEB code, the test centre application and the College Board listing.',
				chips: ['CEEB code', 'Test centre code', 'Listing'],
				option: true,
			},
		],
	},
	partners: {
		heading: 'Schools and universities we work with',
		newTab: 'opens in a new tab',
	},
	stats: {
		heading: 'Work so far',
		items: [
			{ value: '24', label: 'Institutions advised', body: 'Schools, universities and learning centres.' },
			{ value: '6', label: 'Regions of Uzbekistan', body: 'Tashkent, Bukhara, Andijan, Fergana, Samarkand and Kashkadarya.' },
			{ value: '4', label: 'Official SAT test centres', body: 'Schools we took onto the College Board test centre list.' },
		],
	},
	engagement: {
		heading: 'How we work together',
		intro: 'Pricing depends on the institution, but the model is agreed before any work starts.',
		items: [
			{
				title: 'Readiness review',
				price: 'Fixed fee',
				body: 'A short diagnostic of your licence, documents, gaps and timeline. Right when you need a decision before committing.',
			},
			{
				title: 'Licensing project',
				price: 'Fixed scope',
				body: 'End-to-end support for one licensing goal, with evidence tracking and follow-up coordination.',
			},
			{
				title: 'Advisory retainer',
				price: 'Monthly',
				body: 'Ongoing support for schools running several programmes, approvals or launches at once.',
			},
		],
	},
	faq: {
		heading: 'Questions',
		items: [
			{
				question: 'Who do you work with?',
				answer:
					'Private schools preparing for international programme authorisation, universities and education groups building institutional evidence for partnerships, and established institutions adding Cambridge, IB-oriented, American or dual-diploma pathways.',
			},
			{
				question: 'Do you guarantee a licence?',
				answer:
					'No. Exam boards and programme bodies make the final decision. We align you with their requirements and prepare a strong, coherent submission.',
			},
			{
				question: 'Which programmes?',
				answer:
					'Typically British-style pathways, IB-oriented models, American or dual-diploma setups. We confirm fit early, before you commit.',
			},
			{
				question: 'How long does it take?',
				answer:
					'Months, not weeks. The plan follows your starting point, recruitment and target authorisation date.',
			},
			{
				question: 'Can you make us an SAT test centre?',
				answer:
					'Yes, as a separate option. We prepare the CEEB code request, the test centre application and the documents the College Board asks for.',
			},
			{
				question: 'How do we start?',
				answer:
					'Email the institution name, city, current licence status, target programme and timeline. We reply with the right next step rather than a generic package.',
			},
		],
	},
	cta: {
		heading: 'Tell us about your school',
		sub: 'Send your school name, city, licence status and timeline. We reply with the next step.',
		button: 'Email the brief',
		secondaryButton: 'Message on WhatsApp',
		emailSubject: 'Institution assessment request',
		emailBody:
			'Hello Edu License,\n\nInstitution name:\nCity:\nCurrent licence status:\nTarget programme:\nDesired timeline:\nMain contact:\n',
		whatsappPrefill:
			'Hello, we would like an institution assessment. I can share our school name, city, licence status and timeline.',
	},
	footer: {
		tagline: 'International school licensing, setup and launch, from Tashkent.',
		whatsapp: 'WhatsApp',
		email: 'Email',
		telegram: 'Telegram',
		privacy: 'Privacy',
		rights: 'All rights reserved.',
		addressLabel: 'Office',
		address: 'Sayram street, 7th lane, house 21, Mirzo Ulugbek district, Tashkent',
		directions: 'Get directions',
		contactLabel: 'Contact',
		pagesLabel: 'Pages',
	},
	privacy: {
		title: 'Privacy notice',
		body:
			'Edu License uses contact details and institution information only to respond to enquiries, assess project fit, and coordinate agreed work. We do not sell personal data. Documents shared with us are treated as confidential project materials and are used only for the purpose agreed with the institution.',
		back: 'Back to home',
	},
	stateLabels: {
		done: 'Done',
		active: 'In progress',
		next: 'Coming up',
		gap: 'Needs work',
	},
};
