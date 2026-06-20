const TWITTER_HOSTNAMES = new Set(["twitter.com", "x.com", "www.twitter.com", "www.x.com"]);

export const MAX_CLIMB_DEPTH = 40;
const TWEET_TEXT_SELECTOR = '[data-testid="tweetText"]';

export function isTwitterUrl(pageUrl: string): boolean {
  let url: URL;
  try {
    url = new URL(pageUrl);
  } catch {
    return false;
  }
  return TWITTER_HOSTNAMES.has(url.hostname);
}

export function resolveTweetText(target: Element): string | null {
  let current: Element | null = target;
  for (let depth = 0; depth < MAX_CLIMB_DEPTH && current; depth += 1) {
    const match = current.querySelector(TWEET_TEXT_SELECTOR);
    if (match?.textContent) return match.textContent;
    current = current.parentElement;
  }
  return null;
}
