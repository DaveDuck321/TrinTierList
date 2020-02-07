//Returns: {category:{id:int, name:string}, person1:{id:number, name:string, imgs:string[]}, ...}

import { PostJSON, PopulateCategories } from "/common.js";

const kFirst = Symbol("first");
const kSecond = Symbol("second");

const ErrorMessages = {
    "no available matches": `No matches could be found for your user in this category!
Are you permitted to vote?
If you believe this is an error, please contact the administrator.`
};

let RequestedCategory = 0;
let CurrentMatch = {
    id: [],
    category: 0
}

function Sleep(ms) {
    return new Promise(resolve => {
        setTimeout(resolve, ms);
    });
}

function Sample(list) {
    return list[Math.floor(Math.random() * list.length)];
}

function FormatEloChange(Change) {
    let S = Change.toString();

    if (Change >= 0)
        S = "+" + S

    return S;
}

function EloChangeClass(Change) {
    if (Change < 0)
        return "text-danger";
    // (Change >= 0)
    return "text-success";
}

function LoadImage(src) {
    return new Promise(resolve => {
        const image = document.createElement("img");
        image.addEventListener("load", () => resolve(image), { once: true });
        image.src = src;
    });
}

function ReplaceElement(el, newEl) {
    newEl.classList = el.classList;
    el.parentNode.replaceChild(newEl, el);
}

async function ShowPeople(category, AnimationPromise = () => { }) {
    document.getElementById("categoryName").innerHTML = "Loading...";

    const data = await PostJSON("/api/match", { category });

    //Crash out w/ message if non-engineer tries to vote
    if (!data.success) {
        document.getElementById("categoryName").innerHTML = "Error!";
        if (ErrorMessages[data.msg]) {
            alert(ErrorMessages[data.msg]);
        }
        return;
    }

    const old_image_1 = document.querySelector("#first  img");
    const old_image_2 = document.querySelector("#second img");

    const [new_image_1, new_image_2, FinishAnimation] = await Promise.all([
        LoadImage(Sample(data.person1.imgs)),
        LoadImage(Sample(data.person2.imgs)),
        AnimationPromise
    ]);

    FinishAnimation();

    document.getElementById("categoryName").innerHTML = data.category.name;

    ReplaceElement(old_image_1, new_image_1);
    ReplaceElement(old_image_2, new_image_2);

    document.querySelector("#first  h3").innerText = data.person1.nickname;
    document.querySelector("#second h3").innerText = data.person2.nickname;

    document.querySelector("#first  p span:first-of-type").innerText = data.person1.elo;
    document.querySelector("#second p span:first-of-type").innerText = data.person2.elo;

    CurrentMatch = {
        id: [data.person1.id, data.person2.id],
        category: data.category.id,
    };
}

async function AnimateEloChange(elo_change, Winner) {
    if (Winner === kFirst) {
        elo_change.person1 = elo_change.winner;
        elo_change.person2 = elo_change.looser;
    } else {
        elo_change.person1 = elo_change.looser;
        elo_change.person2 = elo_change.winner;
    }

    const Change1 = document.querySelector("#first  p span:last-of-type");
    const Change2 = document.querySelector("#second p span:last-of-type");

    Change1.classList = EloChangeClass(elo_change.person1);
    Change2.classList = EloChangeClass(elo_change.person2);

    Change1.innerText = FormatEloChange(elo_change.person1);
    Change2.innerText = FormatEloChange(elo_change.person2);

    await Sleep(300);

    return () => {
        // Very javascript, much lambda
        Change1.innerText = "";
        Change2.innerText = "";
    }

    /* Removed for now since it makes the site feel jerky -- maybe adjust timings?
    const Elo1 = document.querySelector("#first  p span:first-of-type");
    const Elo2 = document.querySelector("#second p span:first-of-type");

    Elo1.innerText = parseInt(Elo1.innerText) + elo_change.person1;
    Elo2.innerText = parseInt(Elo2.innerText) + elo_change.person2;

    await Sleep(100);*/
}

async function Vote(Winner, category) {
    for (const Button of document.querySelectorAll("button.btn-lg"))
        Button.disabled = true;

    const Data = {
        type: "vote",
        category
    };

    if (Winner === kFirst) {
        Data.won = CurrentMatch.id[0];
        Data.lost = CurrentMatch.id[1];
    } else {
        Data.won = CurrentMatch.id[1];
        Data.lost = CurrentMatch.id[0];
    }

    const Response = await PostJSON("/api/vote", Data);

    const animationFinished = AnimateEloChange(Response.elo_change, Winner);

    await ShowPeople(RequestedCategory, animationFinished);

    for (const Button of document.querySelectorAll("button.btn-lg"))
        Button.disabled = false;
}

document.addEventListener("DOMContentLoaded", async () => {
    document.querySelector("#first  button").onclick = () => { Vote(kFirst, CurrentMatch.category) };
    document.querySelector("#second button").onclick = () => { Vote(kSecond, CurrentMatch.category) };

    const { categories } = await PostJSON("/api/leaderboard");
    PopulateCategories(categories, e => {
        RequestedCategory = parseInt(e.srcElement.value);
        ShowPeople(RequestedCategory);
    });
    ShowPeople(RequestedCategory);
});
