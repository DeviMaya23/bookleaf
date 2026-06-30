import providersData from "./clearUrlsProviders.json";

type ProviderEntry = {
  urlPattern: string;
  completeProvider: boolean;
  rules: string[];
  referralMarketing: string[];
  exceptions: string[];
};

export type CompiledProvider = {
  urlPattern: RegExp;
  completeProvider: boolean;
  rules: RegExp[];
  referralMarketing: RegExp[];
  exceptions: RegExp[];
};

function compileProvider(entry: ProviderEntry): CompiledProvider {
  return {
    urlPattern: new RegExp(entry.urlPattern),
    completeProvider: entry.completeProvider,
    rules: entry.rules.map((r) => new RegExp(`^(?:${r})$`)),
    referralMarketing: entry.referralMarketing.map((r) => new RegExp(`^(?:${r})$`)),
    exceptions: entry.exceptions.map((r) => new RegExp(r)),
  };
}

const compiled: CompiledProvider[] = Object.values(
  (providersData as { providers: Record<string, ProviderEntry> }).providers,
).map(compileProvider);

export function cleanUrl(rawUrl: string, providers: CompiledProvider[] = compiled): string {
  let url: URL;
  try {
    url = new URL(rawUrl);
  } catch {
    return rawUrl;
  }

  for (const provider of providers) {
    if (!provider.urlPattern.test(rawUrl)) continue;
    if (provider.exceptions.some((ex) => ex.test(rawUrl))) continue;

    if (provider.completeProvider) {
      url.search = "";
      return url.toString();
    }

    const stripPatterns = [...provider.rules, ...provider.referralMarketing];
    for (const key of [...url.searchParams.keys()]) {
      if (stripPatterns.some((pattern) => pattern.test(key))) {
        url.searchParams.delete(key);
      }
    }
  }

  return url.toString();
}
