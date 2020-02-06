// {people:[{id:number, name:string, imgs:string[]}], rankings: }

import { PostJSON } from "/common.js";

function sortByCategory(people, elos, category) {
    for (const person of people)
        person.elo = elos[category][person.id];

    people.sort((p1, p2) => p2.elo - p1.elo);
}

async function PopulateLeaderboard(people, elos, category) {
    // Delete existing entries
    document.querySelectorAll("#leaderboard tr").forEach(e => e.remove());

    sortByCategory(people, elos, category);

    const leaderboard = document.getElementById("leaderboard");
    const template = leaderboard.querySelector("template");
    const row = template.content.querySelector("tr");

    for (let [index, person] of people.entries()) {
        let personRow = document.importNode(row, true);

        personRow.querySelector(".rank").innerText = index + 1;
        personRow.querySelector(".nickname").innerText = person.nickname;
        personRow.querySelector(".elo").innerText = person.elo;

        leaderboard.appendChild(personRow);
    }
}

function PopulateCategories(categories) {
    const select = document.getElementById("category");

    for (const category of categories) {
        const option = document.createElement("option");
        option.value = category.id;
        option.innerText = category.name;

        select.appendChild(option);
    }
}

document.addEventListener("DOMContentLoaded", async () => {
    const { people, categories, elos } = await PostJSON("/api/leaderboard", {});

    PopulateCategories(categories);
    PopulateLeaderboard(people, elos, categories[0].id);

    document.getElementById("category").addEventListener("change",
        e => PopulateLeaderboard(people, elos, e.srcElement.value));
});
