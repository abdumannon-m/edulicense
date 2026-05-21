/**
 * Certificate registry.
 *
 * Each entry generates a verification page at `/verify/<slug>` (and
 * `/uz/verify/<slug>`). The QR code printed on a certificate points at that
 * URL. To publish a new certificate, add one record here and rebuild.
 */
export interface Certificate {
	/** URL key, e.g. `oriental-university-sat-center`. */
	slug: string;
	/** Institution / centre name shown as the headline. */
	institution: string;
	/** Role granted, e.g. "Authorised SAT Test Centre". */
	designation: string;
	/** Human-readable SAT administration date. */
	satAdministrationDate: string;
	/** Path (under /public) to the College Board search screenshot. */
	collegeBoardScreenshot: string;
	/** Unique verification identifier printed on the certificate. */
	verificationId: string;
	/** Human-readable issue date. */
	issueDate: string;
}

export const certificates: Certificate[] = [
	{
		slug: 'oriental-university-sat-center',
		institution: 'Oriental University',
		designation: 'Authorised SAT Test Centre',
		satAdministrationDate: '8 March 2025',
		collegeBoardScreenshot: '/certificates/oriental-university-sat-center.svg',
		verificationId: 'EDL-2025-0312-OU',
		issueDate: '12 March 2025',
	},
];

export function getCertificate(slug: string): Certificate | undefined {
	return certificates.find((c) => c.slug === slug);
}
