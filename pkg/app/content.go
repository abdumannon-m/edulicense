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
	Proof           ProofMessages
	Why             WhyMessages
	Process         ProcessMessages
	FAQ             FAQMessages
	Team            TeamMessages
	Engagement      EngagementMessages
	CTABand         CTABandMessages
	Footer          FooterMessages
	Privacy         PrivacyMessages
	Verify          VerifyMessages
}

type NavMessages struct {
	Services   string
	Audience   string
	Proof      string
	Why        string
	Process    string
	Team       string
	Engagement string
	FAQ        string
	Contact    string
	SwitchToEn string
	SwitchToUz string
	Menu       string
	MenuOpen   string
	MenuClose  string
}

type HeroMessages struct {
	Eyebrow         string
	Title           string
	Subtitle        string
	CTA             string
	SecondaryCTA    string
	ScrollHint      string
	WhatsAppPrefill string
	EmailSubject    string
	EmailBody       string
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
	Intro   string
	Items   []string
}

type ProofMessages struct {
	Heading string
	Intro   string
	Stats   []ProofStat
	Items   []ProofItem
}

type ProofStat struct {
	Value string
	Label string
}

type ProofItem struct {
	Title     string
	Body      string
	Href      string
	LinkLabel string
}

type WhyMessages struct {
	Heading string
	Body    string
	Items   []WhyItem
}

type WhyItem struct {
	Title string
	Body  string
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
	Intro   string
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
	SecondaryButton string
	EmailSubject    string
	EmailBody       string
	WhatsAppPrefill string
}

type TeamMessages struct {
	Heading string
	Intro   string
	Items   []TeamItem
}

type TeamItem struct {
	Title string
	Role  string
	Body  string
}

type EngagementMessages struct {
	Heading string
	Intro   string
	Items   []EngagementItem
}

type EngagementItem struct {
	Title string
	Price string
	Body  string
}

type FooterMessages struct {
	Tagline     string
	WhatsApp    string
	Email       string
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
		Proof:      "Proof",
		Why:        "Why us",
		Process:    "Process",
		Team:       "Team",
		Engagement: "Engagement",
		FAQ:        "FAQ",
		Contact:    "Contact us",
		SwitchToEn: "English",
		SwitchToUz: "O'zbekcha",
		Menu:       "Menu",
		MenuOpen:   "Open menu",
		MenuClose:  "Close menu",
	},
	Hero: HeroMessages{
		Eyebrow:         "International school licensing · Tashkent",
		Title:           "Get your school internationally licensed — from zero to launch",
		Subtitle:        "Tashkent-based team helping Uzbekistan schools align with global programmes, stand up strong institutions, and open with confidence.",
		CTA:             "Request an assessment",
		SecondaryCTA:    "WhatsApp us",
		ScrollHint:      "See how we help",
		WhatsAppPrefill: "Hello — I would like to discuss international licensing and school setup for our institution in Uzbekistan.",
		EmailSubject:    "Institution readiness assessment",
		EmailBody:       "Hello Edu License,\n\nWe would like to discuss licensing or SAT test center readiness for our institution.\n\nInstitution name:\nCity:\nCurrent programme:\nTarget timeline:\n",
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
		Intro:   "We are most useful when a school or university needs a practical path from ambition to an evidence-ready application.",
		Items: []string{
			"Private schools preparing for international programme authorisation.",
			"Universities and schools applying to become SAT test centres.",
			"Established institutions adding Cambridge, IB-oriented, American, or dual-diploma pathways.",
			"Leadership teams that need Uzbek context translated into international evidence standards.",
		},
	},
	Proof: ProofMessages{
		Heading: "Evidence, not vague promises",
		Intro:   "Licensing work is high-stakes, so we show the parts that can be verified publicly and keep the rest confidential for each institution.",
		Stats: []ProofStat{
			{Value: "2-step", Label: "SAT centre path: CEEB code, then test centre approval"},
			{Value: "9-point", Label: "website and document readiness checklist before submission"},
			{Value: "Public", Label: "certificate verification record for completed work"},
		},
		Items: []ProofItem{
			{Title: "Oriental University SAT test centre certificate", Body: "The public verification page shows the issued credential and College Board record we can disclose.", Href: "/verify/oriental-university-sat-center", LinkLabel: "View verification"},
			{Title: "Document-first workflow", Body: "We collect the English licence, domain email, English website evidence, responsible staff details, and application ownership before submission."},
			{Title: "Confidential client work", Body: "Many licensing and school-readiness projects cannot be named publicly. References can be discussed during a qualified discovery call."},
		},
	},
	Why: WhyMessages{
		Heading: "Why Edu License",
		Body:    "We combine local execution in Uzbekistan with the paperwork discipline international bodies expect.",
		Items: []WhyItem{
			{Title: "Local evidence control", Body: "We check licence details, website content, addresses, contacts, and operational documents before an application leaves your team."},
			{Title: "Application ownership", Body: "One responsible person tracks the timeline, missing documents, CEEB steps, College Board submission, and follow-up."},
			{Title: "Operational readiness", Body: "The work does not stop at forms. We help align people, schedules, website proof, and communications so the institution can actually operate."},
		},
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
		Intro:   "Short answers to the issues school founders and university teams usually raise before starting.",
		Items: []FAQItem{
			{Question: "Do you guarantee a licence?", Answer: "No — exam boards and programme bodies make final decisions. We align you with requirements and prepare a strong, coherent submission."},
			{Question: "Which programmes?", Answer: "Typically British-style pathways, IB-oriented models, American or dual-diploma setups, and SAT test centre applications. We confirm fit early."},
			{Question: "Timeline?", Answer: "Expect months, not weeks — structured around your maturity, recruitment, and target authorisation date."},
			{Question: "What documents do SAT test centre applicants need first?", Answer: "We start with the English licence, domain email, English website, one responsible staff member, and website proof for leadership, timetable, address, phone, and email."},
			{Question: "How does the SAT test centre process work?", Answer: "It is handled in two main steps: first the institution receives a CEEB code, then the test centre application is submitted and tracked until it appears in the official list."},
			{Question: "How do we start?", Answer: "Send the institution name, city, current licence status, target programme, and desired timeline. We reply with the right next step instead of a generic package."},
		},
	},
	Team: TeamMessages{
		Heading: "A practical team around the application",
		Intro:   "Consulting is delivered by people, not templates. The operating model below shows who is accountable during a project.",
		Items: []TeamItem{
			{Title: "Founder-led advisory", Role: "Strategy and institution fit", Body: "Senior guidance on whether the target licence, programme, or SAT centre path fits the institution before work begins."},
			{Title: "Application operations", Role: "Documents, CEEB, College Board follow-up", Body: "Day-to-day tracking of missing documents, submission steps, responsible staff, reminders, and external communication."},
			{Title: "School readiness coordination", Role: "Website, timetable, contacts, evidence", Body: "Practical support to make public-facing evidence match the licence, address, staff structure, and operational reality."},
		},
	},
	Engagement: EngagementMessages{
		Heading: "Engagement model",
		Intro:   "Exact pricing depends on the institution, but we make the commercial model clear before work starts.",
		Items: []EngagementItem{
			{Title: "Readiness review", Price: "Fixed diagnostic", Body: "A short review of licence, website, documents, gaps, and target timeline. Best when you need a decision before committing to a full project."},
			{Title: "Application project", Price: "Fixed project scope", Body: "End-to-end support for a defined licensing or SAT test centre application, including document tracking and submission follow-up."},
			{Title: "Advisory retainer", Price: "Monthly support", Body: "Ongoing support for schools building multiple pathways, preparing operations, or coordinating several approval tracks at once."},
		},
	},
	CTABand: CTABandMessages{
		Heading:         "Request an institution assessment",
		Sub:             "Share your school name, city, current licence status, and target timeline. We will reply with the right next step.",
		Button:          "Email the brief",
		SecondaryButton: "WhatsApp quick message",
		EmailSubject:    "Institution assessment request",
		EmailBody:       "Hello Edu License,\n\nInstitution name:\nCity:\nCurrent licence status:\nTarget programme or SAT centre goal:\nDesired timeline:\nMain contact:\n",
		WhatsAppPrefill: "Hello — we want to request an institution assessment. I can share our school name, city, licence status, and target timeline.",
	},
	Footer: FooterMessages{
		Tagline:     "Licensing · Setup · Readiness",
		WhatsApp:    "WhatsApp",
		Email:       "Email",
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
		Proof:      "Dalillar",
		Why:        "Nega biz",
		Process:    "Jarayon",
		Team:       "Jamoa",
		Engagement: "Hamkorlik",
		FAQ:        "Savollar",
		Contact:    "Contact us",
		SwitchToEn: "English",
		SwitchToUz: "O'zbekcha",
		Menu:       "Menyu",
		MenuOpen:   "Menyuni ochish",
		MenuClose:  "Menyuni yopish",
	},
	Hero: HeroMessages{
		Eyebrow:         "Xalqaro maktab litsenziyasi · Toshkent",
		Title:           "Maktabingizni xalqaro litsenziyaga — noldan ishga tushguncha",
		Subtitle:        "Toshkentdagi jamoamiz O'zbekiston maktablarini global dasturlarga moslash, mustahkam tuzilma qurish va ishonch bilan ochishda qo'llab-quvvatlaydi.",
		CTA:             "Baholash so'rash",
		SecondaryCTA:    "WhatsApp orqali",
		ScrollHint:      "Qanday yordam berishimiz",
		WhatsAppPrefill: "Assalomu alaykum — O'zbekistondagi maktabimiz uchun xalqaro litsenza va tashkil etish bo'yicha maslahat kerak.",
		EmailSubject:    "Muassasa tayyorgarligini baholash",
		EmailBody:       "Assalomu alaykum Edu License,\n\nMuassasamiz uchun litsenziya yoki SAT test markazi tayyorgarligini muhokama qilmoqchimiz.\n\nMuassasa nomi:\nShahar:\nHozirgi dastur:\nMaqsadli muddat:\n",
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
		Intro:   "Maktab yoki universitet g'oyadan dalillarga tayyor arizagacha aniq yo'lga muhtoj bo'lsa, biz eng foydali bo'lamiz.",
		Items: []string{
			"Xalqaro dastur vakolatini olishga tayyorlanayotgan xususiy maktablar.",
			"SAT test markazi bo'lish uchun ariza topshirayotgan maktab va universitetlar.",
			"Cambridge, IBga yaqin, American yoki dual-diploma yo'nalishlarini qo'shayotgan muassasalar.",
			"O'zbekiston kontekstini xalqaro dalil standartlariga moslashtirishi kerak bo'lgan rahbariyat jamoalari.",
		},
	},
	Proof: ProofMessages{
		Heading: "Umumiy va'dalar emas, dalillar",
		Intro:   "Litsenziya ishlari mas'uliyatli. Shuning uchun oshkor qilish mumkin bo'lgan natijalarni ko'rsatamiz, qolgan mijoz ishlari esa maxfiy qoladi.",
		Stats: []ProofStat{
			{Value: "2 bosqich", Label: "SAT markazi yo'li: avval CEEB kodi, keyin test markazi tasdig'i"},
			{Value: "9 band", Label: "arizadan oldingi website va hujjat tayyorgarligi tekshiruvi"},
			{Value: "Ochiq", Label: "yakunlangan ish uchun sertifikat tekshiruv sahifasi"},
		},
		Items: []ProofItem{
			{Title: "Oriental University SAT test markazi sertifikati", Body: "Ochiq tekshiruv sahifasida berilgan hujjat va ko'rsatish mumkin bo'lgan College Board yozuvi bor.", Href: "/uz/verify/oriental-university-sat-center", LinkLabel: "Tekshiruvni ko'rish"},
			{Title: "Hujjatdan boshlanadigan jarayon", Body: "Arizadan oldin ingliz tilidagi litsenziya, domen email, inglizcha website, mas'ul xodim va website dalillarini yig'amiz."},
			{Title: "Maxfiy mijoz ishlari", Body: "Ko'p litsenziya va tayyorgarlik loyihalarini ochiq nomlash mumkin emas. Tavsiyalar malakali suhbatda muhokama qilinadi."},
		},
	},
	Why: WhyMessages{
		Heading: "Nega Edu License",
		Body:    "Biz O'zbekistondagi mahalliy ijroni xalqaro tashkilotlar kutadigan hujjat intizomi bilan birlashtiramiz.",
		Items: []WhyItem{
			{Title: "Mahalliy dalil nazorati", Body: "Ariza yuborilishidan oldin litsenziya ma'lumotlari, website, manzil, kontaktlar va operatsion hujjatlarni tekshiramiz."},
			{Title: "Ariza egasi aniq", Body: "Bitta mas'ul shaxs timeline, yetishmayotgan hujjatlar, CEEB bosqichlari, College Board arizasi va follow-upni kuzatadi."},
			{Title: "Operatsion tayyorgarlik", Body: "Ish faqat forma to'ldirish bilan tugamaydi. Jamoa, jadval, website dalillari va kommunikatsiyalarni ham moslaymiz."},
		},
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
		Intro:   "Maktab asoschilari va universitet jamoalari boshlashdan oldin ko'p so'raydigan savollar.",
		Items: []FAQItem{
			{Question: "Litsenziya kafolati bormi?", Answer: "Yo'q — yakuniy qarorni imtihon markazi yoki tashkilot qiladi. Biz talablarga moslashtiramiz va kuchli, izchil topshiriq uchun tayyorlaymiz."},
			{Question: "Qaysi dasturlar?", Answer: "Odatda Britaniya uslubidagi yo'llar, IBga yaqin modellar, American yoki dual-diploma yo'nalishlari va SAT test markazi arizalari. Mosligini boshida aniqlaymiz."},
			{Question: "Qancha vaqt?", Answer: "Bir necha oy — boshlang'ich holat, kadrlar va maqsadli vakolatlash sanasiga bog'liq."},
			{Question: "SAT test markazi arizasi uchun avval nimalar kerak?", Answer: "Ingliz tilidagi litsenziya, domen email, inglizcha website, bitta mas'ul xodim va websiteda rahbar, jadval, manzil, telefon hamda email dalillari kerak."},
			{Question: "SAT test markazi jarayoni qanday ishlaydi?", Answer: "Jarayon ikki asosiy bosqichdan iborat: avval muassasa CEEB kodi oladi, keyin test markazi arizasi topshirilib, rasmiy ro'yxatga chiqquncha kuzatiladi."},
			{Question: "Qanday boshlaymiz?", Answer: "Muassasa nomi, shahar, hozirgi litsenziya holati, maqsadli dastur va muddatni yuboring. Biz umumiy paket emas, to'g'ri keyingi qadamni aytamiz."},
		},
	},
	Team: TeamMessages{
		Heading: "Ariza atrofidagi amaliy jamoa",
		Intro:   "Konsalting shablonlar bilan emas, odamlar bilan bajariladi. Quyidagi model loyiha davomida kim mas'ul ekanini ko'rsatadi.",
		Items: []TeamItem{
			{Title: "Founder-led advisory", Role: "Strategiya va muassasa mosligi", Body: "Ish boshlanishidan oldin maqsadli litsenziya, dastur yoki SAT markazi yo'li muassasaga mosligini baholash."},
			{Title: "Application operations", Role: "Hujjatlar, CEEB, College Board follow-up", Body: "Yetishmayotgan hujjatlar, ariza bosqichlari, mas'ul xodimlar, eslatmalar va tashqi kommunikatsiyani kundalik kuzatish."},
			{Title: "School readiness coordination", Role: "Website, jadval, kontaktlar, dalillar", Body: "Ochiq dalillar litsenziya, manzil, xodimlar tuzilmasi va haqiqiy operatsiyaga mos bo'lishini ta'minlash."},
		},
	},
	Engagement: EngagementMessages{
		Heading: "Hamkorlik modeli",
		Intro:   "Aniq narx muassasaga bog'liq, lekin tijoriy model ish boshlanishidan oldin kelishiladi.",
		Items: []EngagementItem{
			{Title: "Tayyorlik tekshiruvi", Price: "Fixed diagnostic", Body: "Litsenziya, website, hujjatlar, bo'shliqlar va timeline bo'yicha qisqa ko'rib chiqish. To'liq loyihadan oldin qaror kerak bo'lsa mos."},
			{Title: "Ariza loyihasi", Price: "Fixed project scope", Body: "Belgilangan litsenziya yoki SAT test markazi arizasi uchun hujjat kuzatuvi va submission follow-up bilan end-to-end yordam."},
			{Title: "Advisory retainer", Price: "Monthly support", Body: "Bir nechta yo'nalish, operatsion tayyorgarlik yoki parallel approval tracklarni yuritayotgan maktablar uchun davomiy yordam."},
		},
	},
	CTABand: CTABandMessages{
		Heading:         "Muassasa baholashini so'rang",
		Sub:             "Maktab nomi, shahar, litsenziya holati va maqsadli muddatni yuboring. Keyingi to'g'ri qadam bilan javob beramiz.",
		Button:          "Email orqali brief",
		SecondaryButton: "WhatsApp tez xabar",
		EmailSubject:    "Muassasa baholashi so'rovi",
		EmailBody:       "Assalomu alaykum Edu License,\n\nMuassasa nomi:\nShahar:\nHozirgi litsenziya holati:\nMaqsadli dastur yoki SAT markazi:\nKerakli muddat:\nAsosiy kontakt:\n",
		WhatsAppPrefill: "Assalomu alaykum — muassasa baholashini so'ramoqchimiz. Maktab nomi, shahar, litsenziya holati va timeline yubora olaman.",
	},
	Footer: FooterMessages{
		Tagline:     "Litsenziya · Tuzilish · Tayyorgarlik",
		WhatsApp:    "WhatsApp",
		Email:       "Email",
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
