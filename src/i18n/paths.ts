/** Build paired EN/UZ URLs for hreflang from current pathname. */
export function localePairUrls(
	site: URL,
	pathname: string,
): { en: string; uz: string } {
	const path = pathname.replace(/\/$/, '') || '/';
	const isUzRoute = path === '/uz' || path.startsWith('/uz/');

	let enPath: string;
	if (!isUzRoute) {
		enPath = path === '' ? '/' : path;
	} else if (path === '/uz') {
		enPath = '/';
	} else {
		enPath = path.slice('/uz'.length) || '/';
	}

	const uzPath = enPath === '/' ? '/uz' : `/uz${enPath}`;

	return {
		en: new URL(enPath, site).href,
		uz: new URL(uzPath, site).href,
	};
}
