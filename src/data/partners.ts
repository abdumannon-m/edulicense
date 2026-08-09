export interface Partner {
	/** Organisation name — also the logo's alt text. */
	name: string;
	/** Path under /public, e.g. '/partners/acme.png'. SVG or transparent PNG. */
	logo: string;
	/** Optional link to the partner's site. */
	href?: string;
	/**
	 * Tile background. Logos are drawn in full colour on a white tile by default.
	 * Set a dark value for logos whose wordmark is white (they would otherwise be
	 * invisible), or match the colour a logo already carries so the artwork blends
	 * into the tile instead of sitting on a visible square.
	 */
	background?: string;
	/**
	 * True when the logo artwork fills its whole canvas (a coloured square, a
	 * photo-like export). The image then covers the tile edge to edge instead of
	 * floating inside the usual padding.
	 */
	bleed?: boolean;
	/**
	 * Optical size nudge, `1` being the tile's natural fit. Logos exported on a
	 * roomy canvas read smaller than wordmarks that run edge to edge; scale those
	 * up a little so the row looks evenly weighted.
	 */
	scale?: number;
}

/**
 * Partner logos shown in the "Partners" section.
 *
 * To add one:
 *   1. Drop the file in `public/partners/` — SVG preferred, otherwise a PNG with
 *      a transparent background, ~400px wide.
 *   2. Add an entry below, with `background` / `bleed` if the artwork needs it.
 *
 * The section hides itself entirely while this list is empty.
 */
export const PARTNERS: Partner[] = [
	{
		name: 'Oriental Universiteti',
		logo: '/partners/oriental-university.png',
		// White wordmark on a transparent canvas — needs a dark tile to be legible.
		background: '#15211c',
	},
	{
		name: 'Cambridge Unit School',
		logo: '/partners/cambridge-unit-school.png',
	},
	// Waiting on the logo files below — save them under these exact names in
	// `public/partners/`, then uncomment.
	// {
	// 	name: 'Karshi International University',
	// 	logo: '/partners/karshi-international-university.png',
	// },
	// {
	// 	name: 'Salam International School',
	// 	logo: '/partners/salam-international-school.png',
	// 	// Artwork is a blue gradient square; let it fill the tile.
	// 	background: '#123a8c',
	// 	bleed: true,
	// },
	// {
	// 	name: "Yakubov's School",
	// 	logo: '/partners/yakubovs-school.png',
	// },
];
