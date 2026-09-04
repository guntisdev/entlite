import { createClient } from "@connectrpc/connect";
import { createConnectTransport } from "@connectrpc/connect-web";
import { timestampDate, timestampFromDate } from "@bufbuild/protobuf/wkt";
import { ArticleService } from "../../ent/gen/ts/schema_pb.js";
import type {
    Article,
    CreateArticleRequest,
    ListArticleFilterByAuthorIsFeaturedPublishedAtTitleRequest,
    UpdateArticleRequest,
} from "../../ent/gen/ts/schema_pb.js";
import {
    checkboxInput,
    createHash,
    daysAgo,
    fakeCoverImage,
    numberInput,
    randomAuthor,
    randomSubtitle,
    randomTitle,
    slugFor,
    textInput,
    toString,
} from "./utils.js";

type StrictMessageInput<T extends { $typeName: string; $unknown?: unknown }> = Omit<T, "$typeName" | "$unknown">;

const transport = createConnectTransport({
    baseUrl: "http://localhost:8080",
});

const articleClient = createClient(ArticleService, transport);

// optional fields of Article, used to report the unset ones
const OPTIONAL_FIELDS = [
    "subtitle", "readingMinutes", "lastViewedMs", "rating", "coverImage", "publishedAt", "metadata",
] as const;

function log(message: string, data?: any) {
    console.log(message, data);
    const output = document.getElementById("output")!;
    const line = document.createElement("div");
    line.textContent = data !== undefined ? `${message} ${toString(data)}` : message;
    output.appendChild(line);
    output.scrollTop = output.scrollHeight;
}

// prefills the id inputs with the last created ID
function rememberID(article: Article) {
    for (const id of ["getArticleId", "updateArticleId", "deleteArticleId"]) {
        (document.getElementById(id) as HTMLInputElement).value = article.ID;
    }
}

function describeArticle(article: Article): string {
    const publishedAt = article.publishedAt ? timestampDate(article.publishedAt).toISOString() : "draft";
    return `ID: ${article.ID} ${article.slug} "${article.title}" by ${article.author} | ${publishedAt}`;
}

// unset optional fields are absent, not zero values
function describeUnset(article: Article): string {
    const unset = OPTIONAL_FIELDS.filter((name) => article[name] === undefined);
    return unset.length > 0 ? `unset (NULL in sqlite): ${unset.join(", ")}` : "every optional field is set";
}

function logArticle(message: string, article: Article) {
    log(message, article);
    log(describeUnset(article));
    rememberID(article);
}

// --- Create ----------------------------------------------------------------

function createFull() {
    const title = randomTitle();
    log("Creating article with every optional field set...");
    const request: StrictMessageInput<CreateArticleRequest> = {
        slug: slugFor(title),
        title: title,
        author: randomAuthor(),
        subtitle: randomSubtitle(),
        readingMinutes: 1 + Math.floor(Math.random() * 20),
        lastViewedMs: BigInt(Date.now()),
        rating: Math.round(Math.random() * 50) / 10,
        coverImage: fakeCoverImage(),
        publishedAt: timestampFromDate(daysAgo(Math.random() * 30)),
        metadata: JSON.stringify({ og_image: "/cover.png", tags: ["entlite", "sqlite"] }),
        isFeatured: true,
    };
    articleClient.createArticle(request)
        .then((response) => {
            logArticle("✓ Article created:", response);
        })
        .catch((error) => {
            log("✗ Error creating article:", error);
        });
}

function createMinimal() {
    const title = randomTitle();
    log("Creating article with only the required fields...");
    // omitted optional fields stay NULL, is_featured takes its default
    const request: StrictMessageInput<CreateArticleRequest> = {
        slug: slugFor(title),
        title: title,
        author: randomAuthor(),
    };
    articleClient.createArticle(request)
        .then((response) => {
            logArticle("✓ Article created:", response);
        })
        .catch((error) => {
            log("✗ Error creating article:", error);
        });
}

function createWithMetadata() {
    const metadata = textInput("createMetadata");
    const title = randomTitle();
    log(`Creating article with metadata ${metadata}...`);
    const request: StrictMessageInput<CreateArticleRequest> = {
        slug: slugFor(title),
        title: title,
        author: randomAuthor(),
        metadata: metadata,
    };
    articleClient.createArticle(request)
        .then((response) => {
            logArticle("✓ Article created:", response);
        })
        .catch((error) => {
            log("✗ Rejected, metadata is not valid json:", error);
        });
}

function createInvalidJSON() {
    const title = randomTitle();
    log("Creating article with metadata '{not json' (checked by the json field validator)...");
    const request: StrictMessageInput<CreateArticleRequest> = {
        slug: slugFor(title),
        title: title,
        author: randomAuthor(),
        metadata: "{not json",
    };
    articleClient.createArticle(request)
        .then((response) => {
            logArticle("✓ Article created (unexpected):", response);
        })
        .catch((error) => {
            log("✗ Rejected by the Validate() interceptor:", error);
        });
}

function createBlankTitle() {
    log("Creating article with a blank title (rejected by logic.NotBlank)...");
    const request: StrictMessageInput<CreateArticleRequest> = {
        slug: slugFor("blank"),
        title: "   ",
        author: randomAuthor(),
    };
    articleClient.createArticle(request)
        .then((response) => {
            logArticle("✓ Article created (unexpected):", response);
        })
        .catch((error) => {
            log("✗ Rejected by the Validate() interceptor:", error);
        });
}

// --- Read ------------------------------------------------------------------

function getByID() {
    const id = textInput("getArticleId");
    if (id === "") {
        log("✗ Invalid article ID");
        return;
    }
    log(`Getting article ${id}...`);
    articleClient.getArticleByID({ ID: id })
        .then((response) => {
            logArticle("✓ Article retrieved:", response);
        })
        .catch((error) => {
            log("✗ Error getting article:", error);
        });
}

function getBySlug() {
    const slug = textInput("getArticleSlug");
    if (slug === "") {
        log("✗ Invalid article slug");
        return;
    }
    log(`Getting article by slug ${slug}...`);
    articleClient.getArticleBySlug({ slug: slug })
        .then((response) => {
            logArticle("✓ Article retrieved:", response);
        })
        .catch((error) => {
            log("✗ Error getting article by slug:", error);
        });
}

// --- Update ----------------------------------------------------------------

// Update needs the whole entity, so read it first
function updateArticle() {
    const id = textInput("updateArticleId");
    if (id === "") {
        log("✗ Invalid article ID");
        return;
    }
    log(`Updating article ${id}...`);
    articleClient.getArticleByID({ ID: id })
        .then((article) => {
            const request: StrictMessageInput<UpdateArticleRequest> = {
                ID: article.ID,
                slug: article.slug,
                title: `${article.title} (${createHash(3)})`,
                author: article.author,
                subtitle: randomSubtitle(),
                readingMinutes: article.readingMinutes,
                lastViewedMs: BigInt(Date.now()),
                rating: article.rating,
                coverImage: article.coverImage,
                publishedAt: article.publishedAt ?? timestampFromDate(new Date()),
                metadata: JSON.stringify({ updated: true, hash: createHash(6) }),
                isFeatured: !article.isFeatured,
            };
            return articleClient.updateArticle(request);
        })
        .then((response) => {
            logArticle("✓ Article updated:", response);
        })
        .catch((error) => {
            // TODO db.UpdateArticleParams has no ID field, so no row matches
            log("✗ Error updating article (known Update ID gap):", error);
        });
}

// omitting optional fields on update sets them back to NULL
function clearOptionals() {
    const id = textInput("updateArticleId");
    if (id === "") {
        log("✗ Invalid article ID");
        return;
    }
    log(`Clearing optional fields of article ${id}...`);
    articleClient.getArticleByID({ ID: id })
        .then((article) => {
            const request: StrictMessageInput<UpdateArticleRequest> = {
                ID: article.ID,
                slug: article.slug,
                title: article.title,
                author: article.author,
            };
            return articleClient.updateArticle(request);
        })
        .then((response) => {
            logArticle("✓ Optional fields cleared:", response);
        })
        .catch((error) => {
            // TODO db.UpdateArticleParams has no ID field, so no row matches
            log("✗ Error clearing optional fields (known Update ID gap):", error);
        });
}

// --- Delete ----------------------------------------------------------------

function deleteArticle() {
    const id = textInput("deleteArticleId");
    if (id === "") {
        log("✗ Invalid article ID");
        return;
    }
    log(`Deleting article ${id}...`);
    articleClient.deleteArticle({ ID: id })
        .then((response) => {
            log("✓ Article deleted:", response);
        })
        .catch((error) => {
            log("✗ Error deleting article:", error);
        });
}

// --- List and filter -------------------------------------------------------

function listAll() {
    log("Listing all articles...");
    articleClient.listAllArticle({})
        .then((response) => {
            log(`✓ Articles listed (${response.rows.length} articles):`);
            response.rows.forEach((article) => log(describeArticle(article)));
        })
        .catch((error) => {
            log("✗ Error listing all articles:", error);
        });
}

function listByAuthor() {
    const author = textInput("listAuthor");
    if (author === "") {
        log("✗ Invalid author");
        return;
    }
    log(`Listing articles of ${author}...`);
    articleClient.listArticleByAuthor({ limit: 50, offset: 0, author: author })
        .then((response) => {
            log(`✓ Articles listed (${response.rows.length} articles):`);
            response.rows.forEach((article) => log(describeArticle(article)));
        })
        .catch((error) => {
            log("✗ Error listing articles:", error);
        });
}

function filterArticles() {
    const author = textInput("filterAuthor");
    if (author === "") {
        log("✗ Invalid author");
        return;
    }
    // only the checked facets are sent
    const request: StrictMessageInput<ListArticleFilterByAuthorIsFeaturedPublishedAtTitleRequest> = {
        limit: 50,
        offset: 0,
        author: author,
    };
    if (checkboxInput("useFeatured")) {
        request.isFeatured = checkboxInput("filterFeatured");
    }
    if (checkboxInput("useTitle")) {
        request.title = textInput("filterTitle");
    }
    if (checkboxInput("useRange")) {
        const days = numberInput("filterDays");
        request.minPublishedAt = timestampFromDate(daysAgo(isNaN(days) ? 30 : days));
        request.maxPublishedAt = timestampFromDate(new Date());
    }
    log("Filtering articles:", request);
    articleClient.listArticleFilterByAuthorIsFeaturedPublishedAtTitle(request)
        .then((response) => {
            log(`✓ Articles filtered (${response.rows.length} articles):`);
            response.rows.forEach((article) => log(describeArticle(article)));
        })
        .catch((error) => {
            // TODO the generated params miss the optional published_at range
            log("✗ Error filtering articles (known filter.Range gap):", error);
        });
}

document.addEventListener("DOMContentLoaded", () => {
    document.getElementById("createFullBtn")!.addEventListener("click", createFull);
    document.getElementById("createMinimalBtn")!.addEventListener("click", createMinimal);
    document.getElementById("createMetadataBtn")!.addEventListener("click", createWithMetadata);
    document.getElementById("createInvalidJsonBtn")!.addEventListener("click", createInvalidJSON);
    document.getElementById("createBlankTitleBtn")!.addEventListener("click", createBlankTitle);

    document.getElementById("getArticleBtn")!.addEventListener("click", getByID);
    document.getElementById("getArticleSlugBtn")!.addEventListener("click", getBySlug);

    document.getElementById("updateArticleBtn")!.addEventListener("click", updateArticle);
    document.getElementById("clearOptionalsBtn")!.addEventListener("click", clearOptionals);
    document.getElementById("deleteArticleBtn")!.addEventListener("click", deleteArticle);

    document.getElementById("listAllBtn")!.addEventListener("click", listAll);
    document.getElementById("listAuthorBtn")!.addEventListener("click", listByAuthor);
    document.getElementById("filterArticleBtn")!.addEventListener("click", filterArticles);

    document.getElementById("clearBtn")!.addEventListener("click", () => {
        document.getElementById("output")!.innerHTML = "";
    });

    log("Entlite Optional Demo Ready!");
    log("Create Full sets every optional field, Create Minimal leaves them NULL");
});
