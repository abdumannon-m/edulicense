package app

type Messages struct {
	Lang            string
	OGLocale        string
	MetaTitle       string
	MetaDescription string
	BrandShort      string
	BrandLegal      string
	Nav             NavMessages
	Hero            HeroMessages
	Pillars         PillarMessages
	Audience        AudienceMessages
	Why             WhyMessages
	Process         ProcessMessages
	FAQ             FAQMessages
	CTABand         CTABandMessages
	Footer          FooterMessages
	Privacy         PrivacyMessages
	Verify          VerifyMessages
}

type NavMessages struct {
	Services   string
	Audience   string
	Why        string
	Process    string
	FAQ        string
	Contact    string
	SwitchToEn string
	SwitchToUz string
	MenuOpen   string
}

type HeroMessages struct {
	Eyebrow         string
	Title           string
	Subtitle        string
	CTA             string
	ScrollHint      string
	WhatsAppPrefill string
	ImageAlt        string
}

type PillarMessages struct {
	Heading string
	Intro   string
	Items   []PillarItem
}

type PillarItem struct {
	Title   string
	Tagline string
	Icon    string
}

type AudienceMessages struct {
	Heading string
	Items   []string
}

type WhyMessages struct {
	Heading string
	Body    string
}

type ProcessMessages struct {
	Heading string
	Steps   []ProcessStep
}

type ProcessStep struct {
	Title string
	Body  string
}

type FAQMessages struct {
	Heading string
	Items   []FAQItem
}

type FAQItem struct {
	Question string
	Answer   string
}

type CTABandMessages struct {
	Heading         string
	Sub             string
	Button          string
	WhatsAppPrefill string
}

type FooterMessages struct {
	Tagline     string
	WhatsApp    string
	Telegram    string
	Privacy     string
	Rights      string
	AddressLine string
	CreditPhoto string
}

type PrivacyMessages struct {
	Title string
	Body  string
	Back  string
}

type VerifyMessages struct {
	MetaTitle         string
	Badge             string
	Heading           string
	Intro             string
	SATDate           string
	VerificationID    string
	IssueDate         string
	ScreenshotHeading string
	ScreenshotCaption string
	SealTop           string
	SealBottom        string
	SignatureLabel    string
	FooterNote        string
	BackHome          string
}

var EnglishMessages = Messages{
	Lang:            "en",
	OGLocale:        "en_GB",
	MetaTitle:       "Edu License LLC · International school licensing · Tashkent",
	MetaDescription: "From zero to launch: international programme licensing, institutional setup, and operational readiness for schools across Uzbekistan.",
	BrandShort:      "Edu License",
	BrandLegal:      "Edu License LLC",
	Nav: NavMessages{
		Services:   "Services",
		Audience:   "Who we help",
		Why:        "Why us",
		Process:    "Process",
		FAQ:        "FAQ",
		Contact:    "Contact us",
		SwitchToEn: "English",
		SwitchToUz: "O'zbekcha",
		MenuOpen:   "Open menu",
	},
	Hero: HeroMessages{
		Eyebrow:         "International school licensing · Tashkent",
		Title:           "Get your school internationally licensed — from zero to launch",
		Subtitle:        "Tashkent-based team helping Uzbekistan schools align with global programmes, stand up strong institutions, and open with confidence.",
		CTA:             "Contact us",
		ScrollHint:      "See how we help",
		WhatsAppPrefill: "Hello — I would like to discuss international licensing and school setup for our institution in Uzbekistan.",
		ImageAlt:        "Students learning together in a bright classroom",
	},
	Pillars: PillarMessages{
		Heading: "Three ways we help",
		Items: []PillarItem{
			{Title: "Licensing", Tagline: "Map your path to authorisation with the frameworks that fit your school — fewer gaps, clearer evidence.", Icon: "license"},
			{Title: "Institutional setup", Tagline: "Governance, handbooks, and academic systems that match what inspectors and partners expect.", Icon: "institution"},
			{Title: "Operational readiness", Tagline: "Roles, timetables, and launch rhythms so your team is ready for day one — not just paperwork.", Icon: "operations"},
		},
	},
	Audience: AudienceMessages{
		Heading: "Who we work with",
		Items: []string{
			"New schools pursuing an international licence.",
			"Established schools adding or changing a global programme.",
		},
	},
	Why: WhyMessages{
		Heading: "Why Edu License",
		Body:    "We are in Tashkent and work across Uzbekistan — local presence, international standards, and a single thread from licensing through opening day.",
	},
	Process: ProcessMessages{
		Heading: "How it works",
		Steps: []ProcessStep{
			{Title: "Discovery", Body: "Goals, your programme, constraints, and timeline — aligned in one session."},
			{Title: "Roadmap", Body: "A sequenced plan you can execute: people, documents, facilities."},
			{Title: "Build", Body: "We work beside your leadership until the model is real — not theoretical."},
			{Title: "Launch", Body: "Handoff with playbooks and optional follow-on as you grow."},
		},
	},
	FAQ: FAQMessages{
		Heading: "Questions",
		Items: []FAQItem{
			{Question: "Do you guarantee a licence?", Answer: "No — exam boards and programme bodies make final decisions. We align you with requirements and prepare a strong, coherent submission."},
			{Question: "Which programmes?", Answer: "Typically British-style pathways, IB-oriented models, and dual-diploma setups. We confirm fit early."},
			{Question: "Timeline?", Answer: "Expect months, not weeks — structured around your maturity, recruitment, and target authorisation date."},
		},
	},
	CTABand: CTABandMessages{
		Heading:         "Ready when you are",
		Sub:             "One message starts the conversation — we reply on WhatsApp.",
		Button:          "Contact us",
		WhatsAppPrefill: "Hello — we want to explore international licensing. What are the next steps?",
	},
	Footer: FooterMessages{
		Tagline:     "Licensing · Setup · Readiness",
		WhatsApp:    "WhatsApp",
		Telegram:    "Telegram",
		Privacy:     "Privacy",
		Rights:      "All rights reserved.",
		AddressLine: "Tashkent, Uzbekistan",
		CreditPhoto: "Classroom photo: Unsplash",
	},
	Privacy: PrivacyMessages{
		Title: "Privacy notice (stub)",
		Body:  "Placeholder only. Add your real policy before collecting personal data beyond WhatsApp messages.",
		Back:  "Back to home",
	},
	Verify: VerifyMessages{
		MetaTitle:         "Certificate verification",
		Badge:             "Verified",
		Heading:           "This certificate is authentic",
		Intro:             "The credential below was issued and verified by Edu License LLC. Details match our records.",
		SATDate:           "SAT administration date",
		VerificationID:    "Verification ID",
		IssueDate:         "Issue date",
		ScreenshotHeading: "College Board record",
		ScreenshotCaption: "Screenshot from the official College Board test centre search.",
		SealTop:           "EDU LICENSE LLC",
		SealBottom:        "VERIFIED CREDENTIAL",
		SignatureLabel:    "Authorised signature",
		FooterNote:        "If any detail does not match the printed certificate, contact Edu License LLC before relying on this document.",
		BackHome:          "Back to Edu License",
	},
}

var UzbekMessages = Messages{
	Lang:            "uz",
	OGLocale:        "uz_UZ",
	MetaTitle:       "Edu License LLC · xalqaro maktab litsenziyasi · Toshkent",
	MetaDescription: "Noldan ishga tushguncha: xalqaro dastur litsenziyasi, tashkiliy tuzilish va tayyorgarlik — O'zbekiston maktablari uchun.",
	BrandShort:      "Edu License",
	BrandLegal:      "Edu License LLC",
	Nav: NavMessages{
		Services:   "Xizmatlar",
		Audience:   "Kimlarga",
		Why:        "Nega biz",
		Process:    "Jarayon",
		FAQ:        "Savollar",
		Contact:    "Contact us",
		SwitchToEn: "English",
		SwitchToUz: "O'zbekcha",
		MenuOpen:   "Menyuni ochish",
	},
	Hero: HeroMessages{
		Eyebrow:         "Xalqaro maktab litsenziyasi · Toshkent",
		Title:           "Maktabingizni xalqaro litsenziyaga — noldan ishga tushguncha",
		Subtitle:        "Toshkentdagi jamoamiz O'zbekiston maktablarini global dasturlarga moslash, mustahkam tuzilma qurish va ishonch bilan ochishda qo'llab-quvvatlaydi.",
		CTA:             "Contact us",
		ScrollHint:      "Qanday yordam berishimiz",
		WhatsAppPrefill: "Assalomu alaykum — O'zbekistondagi maktabimiz uchun xalqaro litsenza va tashkil etish bo'yicha maslahat kerak.",
		ImageAlt:        "Yorqin sinfda birga o'qiyotgan o'quvchilar",
	},
	Pillars: PillarMessages{
		Heading: "Uch yo'nalishda yon bo'lamiz",
		Items: []PillarItem{
			{Title: "Litsenziya", Tagline: "Maktabingizga mos ramkalar bo'yicha vakolatlash yo'lini aniqlaymiz — kamroq bo'shliq, aniq dalillar.", Icon: "license"},
			{Title: "Tashkiliy tuzilish", Tagline: "Boshqaruv, qo'llanmalar va o'quv tizimi — tekshiruv va hamkorlar kutilgan darajada.", Icon: "institution"},
			{Title: "Operatsion tayyorlik", Tagline: "Rollar, jadval va ishga tushirish ritmi — faqat hujjat emas, birinchi kunga tayyor jamoa.", Icon: "operations"},
		},
	},
	Audience: AudienceMessages{
		Heading: "Kim bilan ishlaymiz",
		Items: []string{
			"Xalqaro litsenziya qidirayotgan yangi maktablar.",
			"Global dasturni qo'shmoqchi yoki almashtirmoqchi maktablar.",
		},
	},
	Why: WhyMessages{
		Heading: "Nega Edu License",
		Body:    "Biz Toshkentdamiz va butun O'zbekistonda ishlaymiz — mahalliy mavjudlik, xalqaro standart va litsenziyadan ochilish kunigacha bitta yo'l.",
	},
	Process: ProcessMessages{
		Heading: "Qanday ishlaydi",
		Steps: []ProcessStep{
			{Title: "Tahlil", Body: "Maqsad, dastur, cheklovlar va muddat — bir maromda aniqlanadi."},
			{Title: "Yo'l xaritasi", Body: "Odamlar, hujjatlar, infratuzilma — ketma-ket bajariladigan rejalar."},
			{Title: "Qurish", Body: "Rahbariyat yoningizda — model amaliy bo'lguncha."},
			{Title: "Ishga tushirish", Body: "Qo'llanmalar va kerak bo'lsa keyingi bosqichda maslahat."},
		},
	},
	FAQ: FAQMessages{
		Heading: "Savollar",
		Items: []FAQItem{
			{Question: "Litsenziya kafolati bormi?", Answer: "Yo'q — yakuniy qarorni imtihon markazi yoki tashkilot qiladi. Biz talablarga moslashtiramiz va kuchli, izchil topshiriq uchun tayyorlaymiz."},
			{Question: "Qaysi dasturlar?", Answer: "Odatda Britaniya uslubidagi yo'llar, IBga yaqin modellar va ikki diplom. Mosligini bosqichda aniqlaymiz."},
			{Question: "Qancha vaqt?", Answer: "Bir necha oy — boshlang'ich holat, kadrlar va maqsadli vakolatlash sanasiga bog'liq."},
		},
	},
	CTABand: CTABandMessages{
		Heading:         "Tayyor bo'lsangiz",
		Sub:             "Bitta xabar suhbatni boshlaydi — WhatsApp orqali javob beramiz.",
		Button:          "Contact us",
		WhatsAppPrefill: "Assalomu alaykum — xalqaro litsenziyani o'rganmoqchimiz. Keyingi qadamlar qanday?",
	},
	Footer: FooterMessages{
		Tagline:     "Litsenziya · Tuzilish · Tayyorgarlik",
		WhatsApp:    "WhatsApp",
		Telegram:    "Telegram",
		Privacy:     "Maxfiylik",
		Rights:      "Barcha huquqlar himoyalangan.",
		AddressLine: "Toshkent, O'zbekiston",
		CreditPhoto: "Rasm: Unsplash",
	},
	Privacy: PrivacyMessages{
		Title: "Maxfiylik (stub)",
		Body:  "Vaqtinchina. WhatsAppdan tashqari ma'lumot yig'ishdan oldin haqiqiy siyosatni joylashtiring.",
		Back:  "Bosh sahifa",
	},
	Verify: VerifyMessages{
		MetaTitle:         "Sertifikatni tekshirish",
		Badge:             "Tasdiqlangan",
		Heading:           "Ushbu sertifikat haqiqiy",
		Intro:             "Quyidagi hujjat Edu License LLC tomonidan berilgan va tasdiqlangan. Ma'lumotlar bizning yozuvlarimizga mos keladi.",
		SATDate:           "SAT imtihon sanasi",
		VerificationID:    "Tekshiruv IDsi",
		IssueDate:         "Berilgan sana",
		ScreenshotHeading: "College Board yozuvi",
		ScreenshotCaption: "College Boardning rasmiy test markazlari qidiruvidan olingan skrinshot.",
		SealTop:           "EDU LICENSE LLC",
		SealBottom:        "TASDIQLANGAN HUJJAT",
		SignatureLabel:    "Vakolatli imzo",
		FooterNote:        "Agar biron ma'lumot bosma sertifikatga mos kelmasa, ushbu hujjatga ishonishdan oldin Edu License LLC bilan bog'laning.",
		BackHome:          "Edu License sahifasiga",
	},
}

func MessagesForLocale(locale string) Messages {
	if locale == "uz" {
		return UzbekMessages
	}
	return EnglishMessages
}
