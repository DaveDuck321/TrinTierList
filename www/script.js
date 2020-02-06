//Returns: {category:{id:int, name:string}, person1:{id:number, name:string, imgs:string[]}, ...}

import { PostJSON } from "/common.js";

const kFirst = Symbol("first");
const kSecond = Symbol("second");

let CurrentRank = {
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
    if (Change == 0)
        return "";

    let S = Change.toString();

    if (Change > 0)
        S = "+" + S

    return S;
}

function EloChangeClass(Change) {
    if (Change < 0)
        return "text-danger";
    else if (Change == 0)
        return "text-muted";
    else // (Change > 0)
        return "text-success";
}

async function ShowPeople(category) {
    document.getElementById("category").innerHTML = "Loading...";

    const data = await PostJSON("/api/match", { category });

    document.getElementById("category").innerHTML = data.category.name;

    const image_1 = document.querySelector("#first  img");
    const image_2 = document.querySelector("#second img");

    image_1.src = Sample(data.person1.imgs);
    image_2.src = Sample(data.person2.imgs);

    document.querySelector("#first  h3").innerText = data.person1.nickname;
    document.querySelector("#second h3").innerText = data.person2.nickname;

    document.querySelector("#first  p span:first-of-type").innerText = data.person1.elo;
    document.querySelector("#second p span:first-of-type").innerText = data.person2.elo;

    CurrentRank = {
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

    await Sleep(750);

    Change1.innerText = "";
    Change2.innerText = "";

    Elo1 = document.querySelector("#first  p span:first-of-type");
    Elo2 = document.querySelector("#second p span:first-of-type");

    Elo1.innerText = parseInt(Elo1.innerText) + elo_change.person1;
    Elo2.innerText = parseInt(Elo2.innerText) + elo_change.person2;

    await Sleep(750);
}

async function Vote(Winner, category) {
    for (Button of document.querySelectorAll("button.btn-lg"))
        Button.disabled = true;

    const Data = {
        type: "vote",
        category
    };

    if (Winner === kFirst) {
        Data.won = CurrentRank.id[0];
        Data.lost = CurrentRank.id[1];
    } else {
        Data.won = CurrentRank.id[1];
        Data.lost = CurrentRank.id[0];
    }

    const Response = await PostJSON("/api/vote", Data);

    if (Response.success)
        await AnimateEloChange(Response.elo_change, Winner);

    ShowPeople("random");

    for (Button of document.querySelectorAll("button.btn-lg"))
        Button.disabled = false;
}

document.addEventListener("DOMContentLoaded", () => {
    ShowPeople("random");

    document.querySelector("#first  button").onclick = () => { Vote(kFirst, CurrentRank.category) };
    document.querySelector("#second button").onclick = () => { Vote(kSecond, CurrentRank.category) };
});
