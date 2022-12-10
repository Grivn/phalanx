import matplotlib.pyplot as plt

bsize = [
    ('Safe-Rate', [
        (0, 20.291629),
        (1, 11.170999),
        (2, 6.284638),
        (3, 5.230297),
        (4, 4.139687),
        (5, 2.452470),
        (6, 1.962865),
        (7, 1.578192),
        (8, 2.288075),
        (9, 2.049046),
        (10, 2.278843),
    ], '-o', 'coral'),
    ('Front-Attacked-Rate', [
        (0, 0),
        (1, 0.420679),
        (2, 2.372347),
        (3, 5.580280),
        (4, 14.616175),
        (5, 26.546046),
        (6, 48.450171),
        (7, 55.255443),
        (8, 61.349011),
        (9, 61.863760),
        (10, 67.431887),
    ], '-o', 'brown'),
    ('Front-Attacked-Rate(Risk)', [
        (0, 0),
        (1, 0.420679),
        (2, 2.372347),
        (3, 5.580280),
        (4, 14.419685),
        (5, 26.121987),
        (6, 47.730201),
        (7, 47.730201),
        (8, 60.300310),
        (9, 60.692098),
        (10, 65.842583),
    ], '-p', 'darkseagreen'),
    ('Front-Attacked-Rate(Safe)', [
        (0, 0),
        (1, 0),
        (2, 0),
        (3, 0),
        (4, 0.196490),
        (5, 0.424058),
        (6, 0.719970),
        (7, 0.611316),
        (8, 1.048701),
        (9, 1.589304),
        (10, 1.589304),
    ], '-o', 'gold'),
]


def do_plot():
    f = plt.figure(1, figsize=(8, 4))
    plt.clf()
    ax = f.add_subplot(1, 1, 1)
    for name, entries, style, color in bsize:
        throughput = []
        replica_no = []
        for c, t in entries:
            throughput.append(t)
            replica_no.append(c)
        ax.plot(replica_no, throughput, style, color=color, label='%s' % name, markersize=6, alpha=0.8)
    plt.legend(loc=2, bbox_to_anchor=(1.05, 1.0), borderaxespad=0.)
    plt.grid(linestyle='--', alpha=0.3)
    plt.ylim([0, 100])
    plt.ylabel('Rate (%)')
    plt.xlabel('Number of Byzantine Nodes.')
    plt.tight_layout()
    plt.savefig('front-attack-rate-overload.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
