# Edu License

## What it is
Edu License LLC is a Tashkent team that takes schools and universities in Uzbekistan through international licensing: programme fit and authorisation evidence, institutional setup (governance, handbooks, academic systems) and operational readiness for launch. Authorising a school as an official SAT test centre (CEEB code, test centre code, College Board listing) is one of the services, not the headline.

## Who it is for
Founders and leadership teams of private schools, universities and education groups in Uzbekistan who need Uzbek context translated into international evidence standards. They read in English or Uzbek (Latin), often on a phone, and decide after a conversation, not a checkout.

## Surfaces
- Public landing site (Astro, `src/`), English at `/` and Uzbek at `/uz`. Mode: Persuade. The action is emailing an assessment brief; WhatsApp is the secondary channel.
- Private admin (Go templates, `pkg/templates/`). Mode: Operate.

## Proof we can state
From the production database, October 2026: 24 institutions advised across 6 regions (Tashkent, Bukhara, Andijan, Fergana, Samarkand, Kashkadarya); 4 schools taken to the official SAT test centre list; partner institutions listed in `src/data/partners.ts`.

## Constraints
- No guarantee of a licence: exam boards and programme bodies decide.
- Pricing is quoted per institution; the commercial model is explained before work starts.
- Copy exists in English and Uzbek and must stay in step.
