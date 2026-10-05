import type { Messages } from './types';

/** Uzbek (Latin) copy */
export const uz: Messages = {
	lang: 'uz',
	ogLocale: 'uz_UZ',
	metaTitle: 'Edu License · O‘zbekistonda xalqaro maktab litsenziyasi',
	metaDescription:
		'O‘zbekistondagi maktab va universitetlarni birinchi tayyorgarlik tahlilidan xalqaro vakolatgacha olib boramiz: litsenziya, tashkiliy tuzilma va ishga tushirish.',
	brandShort: 'Edu License',
	brandLegal: 'Edu License LLC',
	nav: {
		how: 'Qanday ishlaydi',
		services: 'Xizmatlar',
		partners: 'Hamkorlar',
		pricing: 'Narxlar',
		faq: 'Savollar',
		contact: 'Email yozish',
		switchToEn: 'English',
		switchToUz: "O'zbekcha",
		menuOpen: 'Menyuni ochish',
		menuClose: 'Menyuni yopish',
		sections: 'Sahifa bo‘limlari',
	},
	hero: {
		title: 'Maktabingiz uchun xalqaro litsenziya',
		subtitle:
			'Toshkentdagi jamoamiz maktab va universitetlarni birinchi tayyorgarlik tahlilidan vakolat olishgacha olib boradi va ochilguningizcha yoningizda bo‘ladi.',
		cta: 'Email orqali yozish',
		secondaryCta: 'WhatsApp orqali yozish',
		emailSubject: 'Muassasa tayyorgarligini baholash',
		emailBody:
			'Assalomu alaykum, Edu License!\n\nMuassasamiz uchun xalqaro litsenziyani muhokama qilmoqchimiz.\n\nMuassasa nomi:\nShahar:\nHozirgi litsenziya holati:\nMaqsadli dastur:\nKerakli muddat:\nAsosiy kontakt:\n',
		whatsappPrefill:
			'Assalomu alaykum, muassasamiz uchun xalqaro litsenziya bo‘yicha maslahatlashmoqchiman.',
	},
	roadmap: {
		title: 'Litsenziya yo‘l xaritasi',
		example: 'Namuna',
		school: 'Sizning maktabingiz · Toshkent',
		progress: '5 bosqichdan 2 tasi bajarildi',
		steps: [
			{ label: 'Tayyorgarlik tahlili', value: '6 ta kamchilik topildi', state: 'done' },
			{ label: 'Dasturni tanlash', value: 'Cambridge yo‘nalishi', state: 'done' },
			{ label: 'Dalil va hujjatlar', value: '20 tadan 12 tasi tayyor', state: 'active' },
			{ label: 'Topshirish', value: 'Mart oyiga rejalangan', state: 'next' },
			{ label: 'Vakolat', value: 'Kengash qarori', state: 'next' },
		],
		nextLabel: 'Keyingi qadam',
		next: 'Boshqaruv qo‘llanmasi loyihasi, juma kuni direktoringiz bilan ko‘rib chiqiladi.',
	},
	how: {
		heading: 'Qanday ishlaydi',
		intro: 'Uch bosqich, butun yo‘l davomida yoningizda bitta jamoa.',
		items: [
			{
				step: '01',
				title: 'Holatingizni tahlil qilamiz',
				body: 'Litsenziya, veb-sayt, siyosatlar va kadrlar bo‘yicha bitta uchrashuv. Kamchiliklar ro‘yxati va sizga nima mosligi haqida ochiq xulosa olasiz.',
				cardTitle: 'Tayyorgarlik tahlili',
				rows: [
					{ label: 'Maktab litsenziyasi', value: 'Joyida', state: 'done' },
					{ label: 'O‘quv dasturi hujjatlari', value: '3 ta kamchilik', state: 'gap' },
					{ label: 'Bolalar xavfsizligi siyosati', value: 'Yo‘q', state: 'gap' },
					{ label: 'O‘qituvchilar malakasi', value: 'Joyida', state: 'done' },
				],
			},
			{
				step: '02',
				title: 'Eng qisqa yo‘lni belgilaymiz',
				body: 'Jamoangiz bajara oladigan ketma-ket reja: kim nima qiladi, qaysi hujjatlar, qachongacha.',
				cardTitle: 'Yo‘l xaritasi',
				rows: [
					{ label: 'Oktyabr', value: 'Boshqaruv va qo‘llanmalar', state: 'active' },
					{ label: 'Dekabr', value: 'Baholash va qabul', state: 'next' },
					{ label: 'Fevral', value: 'Sinov tekshiruvi', state: 'next' },
					{ label: 'Mart', value: 'Topshirish', state: 'next' },
				],
			},
			{
				step: '03',
				title: 'Jamoangiz bilan birga quramiz',
				body: 'Rahbariyatingiz bilan birga yozamiz, tekshiramiz va tayyorlaymiz, dalillar haqiqiy bo‘lguncha. Keyin ishga tushirish uchun qo‘llanmalarni topshiramiz.',
				cardTitle: 'Dalillar to‘plami',
				rows: [
					{ label: 'Boshqaruv qo‘llanmasi', value: 'Tasdiqlandi', state: 'done' },
					{ label: 'Baholash siyosati', value: 'Ko‘rib chiqilmoqda', state: 'active' },
					{ label: 'Xodimlar qo‘llanmasi', value: 'Yozilmoqda', state: 'active' },
					{ label: 'Qabul siyosati', value: 'Navbatda', state: 'next' },
				],
			},
		],
	},
	services: {
		heading: 'Nimalarda yordam beramiz',
		intro: 'Bitta yo‘nalishni yoki butun yo‘lni tanlang. Mosligini birinchi suhbatda aniqlaymiz.',
		optionTag: 'Qo‘shimcha',
		items: [
			{
				title: 'Xalqaro litsenziya',
				body: 'Maktabingizga mos ramka bo‘yicha vakolat olish yo‘lini belgilaymiz: kamroq kamchilik, aniqroq dalillar.',
				chips: ['Cambridge', 'IB yo‘nalishi', 'American', 'Dual diploma'],
			},
			{
				title: 'Tashkiliy tuzilma',
				body: 'Inspektorlar va hamkorlar kutadigan boshqaruv, qo‘llanmalar va o‘quv tizimlari.',
				chips: ['Boshqaruv', 'Qo‘llanmalar', 'O‘quv tizimi'],
			},
			{
				title: 'Ishga tayyorlik',
				body: 'Rollar, dars jadvallari va ishga tushirish rejasi: jamoangiz faqat qog‘ozda emas, birinchi kunga tayyor bo‘ladi.',
				chips: ['Rollar', 'Jadvallar', 'Ishga tushirish'],
			},
			{
				title: 'SAT test markazi',
				body: 'Rasmiy SAT test markaziga aylaning: CEEB kodi, test markazi arizasi va College Board ro‘yxatini biz hal qilamiz.',
				chips: ['CEEB kodi', 'Test markazi kodi', 'Ro‘yxat'],
				option: true,
			},
		],
	},
	partners: {
		heading: 'Biz bilan ishlagan maktab va universitetlar',
		newTab: 'yangi oynada ochiladi',
	},
	stats: {
		heading: 'Natijalarimiz',
		items: [
			{ value: '24', label: 'Muassasaga maslahat berdik', body: 'Maktablar, universitetlar va o‘quv markazlari.' },
			{ value: '6', label: 'O‘zbekiston hududi', body: 'Toshkent, Buxoro, Andijon, Farg‘ona, Samarqand va Qashqadaryo.' },
			{ value: '4', label: 'Rasmiy SAT test markazi', body: 'College Board test markazlari ro‘yxatiga kiritgan maktablarimiz.' },
		],
	},
	engagement: {
		heading: 'Qanday hamkorlik qilamiz',
		intro: 'Narx muassasaga bog‘liq, lekin hamkorlik modeli ish boshlanishidan oldin kelishiladi.',
		items: [
			{
				title: 'Tayyorgarlik tahlili',
				price: 'Belgilangan narx',
				body: 'Litsenziya, hujjatlar, kamchiliklar va muddatlar bo‘yicha qisqa tahlil. Katta loyihadan oldin qaror qabul qilish uchun.',
			},
			{
				title: 'Litsenziya loyihasi',
				price: 'Belgilangan hajm',
				body: 'Bitta litsenziya maqsadi uchun boshidan oxirigacha yordam, dalillar nazorati va kelishuvlar bilan.',
			},
			{
				title: 'Doimiy maslahat',
				price: 'Oylik',
				body: 'Bir vaqtda bir nechta dastur, ruxsatnoma yoki ishga tushirish bilan shug‘ullanayotgan maktablar uchun.',
			},
		],
	},
	faq: {
		heading: 'Savollar',
		items: [
			{
				question: 'Kim bilan ishlaysiz?',
				answer:
					'Xalqaro dastur vakolatiga tayyorlanayotgan xususiy maktablar, hamkorlik uchun institutsional dalillar tayyorlayotgan universitet va ta’lim guruhlari hamda Cambridge, IB yo‘nalishi, American yoki dual-diploma dasturlarini qo‘shayotgan muassasalar bilan.',
			},
			{
				question: 'Litsenziyani kafolatlaysizmi?',
				answer:
					'Yo‘q. Yakuniy qarorni imtihon kengashlari va dastur tashkilotlari qabul qiladi. Biz sizni ularning talablariga moslaymiz va kuchli, izchil hujjatlar to‘plamini tayyorlaymiz.',
			},
			{
				question: 'Qaysi dasturlar?',
				answer:
					'Odatda Britaniya yo‘nalishlari, IB yo‘nalishidagi modellar, American yoki dual-diploma. Mosligini boshidayoq aniqlaymiz.',
			},
			{
				question: 'Qancha vaqt oladi?',
				answer:
					'Haftalar emas, oylar. Reja boshlang‘ich holatingiz, kadrlar va maqsadli vakolat sanasiga qarab tuziladi.',
			},
			{
				question: 'Bizni SAT test markaziga aylantira olasizmi?',
				answer:
					'Ha, alohida xizmat sifatida. CEEB kodi so‘rovi, test markazi arizasi va College Board so‘raydigan hujjatlarni tayyorlaymiz.',
			},
			{
				question: 'Qanday boshlaymiz?',
				answer:
					'Muassasa nomi, shahar, hozirgi litsenziya holati, maqsadli dastur va muddatni emailga yozing. Umumiy paket emas, to‘g‘ri keyingi qadam bilan javob beramiz.',
			},
		],
	},
	cta: {
		heading: 'Maktabingiz haqida yozing',
		sub: 'Maktab nomi, shahar, litsenziya holati va muddatni yuboring. Keyingi qadam bilan javob beramiz.',
		button: 'Email orqali yozish',
		secondaryButton: 'WhatsApp orqali yozish',
		emailSubject: 'Muassasani baholash so‘rovi',
		emailBody:
			'Assalomu alaykum, Edu License!\n\nMuassasa nomi:\nShahar:\nHozirgi litsenziya holati:\nMaqsadli dastur:\nKerakli muddat:\nAsosiy kontakt:\n',
		whatsappPrefill:
			'Assalomu alaykum, muassasamizni baholab berishingizni so‘ramoqchimiz. Maktab nomi, shahar, litsenziya holati va muddatni yubora olaman.',
	},
	footer: {
		tagline: 'Xalqaro maktab litsenziyasi, tashkiliy tuzilma va ishga tushirish. Toshkentdan.',
		whatsapp: 'WhatsApp',
		email: 'Email',
		telegram: 'Telegram',
		privacy: 'Maxfiylik',
		rights: 'Barcha huquqlar himoyalangan.',
		addressLabel: 'Ofis',
		address: 'Toshkent, Mirzo Ulug‘bek tumani, Sayram ko‘chasi, 7-tor, 21-uy',
		directions: 'Yo‘lni ko‘rsatish',
		contactLabel: 'Aloqa',
		pagesLabel: 'Sahifalar',
	},
	privacy: {
		title: 'Maxfiylik siyosati',
		body:
			'Edu License kontakt ma’lumotlari va muassasa haqidagi ma’lumotlardan faqat so‘rovlarga javob berish, loyiha mosligini baholash va kelishilgan ishlarni muvofiqlashtirish uchun foydalanadi. Shaxsiy ma’lumotlarni sotmaymiz. Bizga yuborilgan hujjatlar maxfiy loyiha materiallari sifatida ko‘riladi va faqat muassasa bilan kelishilgan maqsadda ishlatiladi.',
		back: 'Bosh sahifa',
	},
	stateLabels: {
		done: 'Bajarildi',
		active: 'Jarayonda',
		next: 'Navbatda',
		gap: 'Ishlash kerak',
	},
};
