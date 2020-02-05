// {people:[{id:number, name:string, imgs:string[]}], rankings: }


async function PostJSON(url, data) {
    const result = await fetch(url, {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json;'
        },
        body: JSON.stringify(data),
    });
    const response = await result.json();
    if (!response.success) {
        console.error(response.msg);
    }
    return response;
}

function sortByCategory(people, elos, category) {
    for(let person of people) {
        person.elo = elos[category][person.id];
    }

    people.sort((p1, p2) => p2.elo - p1.elo );
}

async function PopulateLeaderboard(people, elos, categories, category) {
    sortByCategory(people, elos, category);
    document.getElementById("category").innerText = categories[category].name;

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

document.addEventListener("DOMContentLoaded", async ()=> {
    let { people, categories, elos } = await PostJSON("/api/leaderboard", {});
    PopulateLeaderboard(people, elos, categories, categories[0].id);
});
