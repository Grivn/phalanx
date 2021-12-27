import matplotlib.pyplot as plt

bsize = [
    ('Safe-Rate', [
        (0, 80.53),
        (1, 67.317767),
        (2, 48.296150),
        (3, 26.824851),
        (4, 16.280314),
        (5, 10.566827),
        (6, 7.110083),
        (7, 8.284638),
        (8, 7.230297),
        (9, 6.139687),
        (10, 5.245343),
    ], '-o', 'coral'),
    ('Front-Attacked-Rate', [
        (0, 0),
        (1, 0.000000),
        (2, 0.000000),
        (3, 0.000000),
        (4, 1.973371),
        (5, 3.240367),
        (6, 9.788574),
        (7, 19.255443),
        (8, 40.349011),
        (9, 61.863760),
        (10, 73.116621),
    ], '-o', 'brown'),
    ('Front-Attacked-Rate(Risk)', [
        (0, 0),
        (1, 0),
        (2, 0),
        (3, 0),
        (4, 1.419685),
        (5, 3.121987),
        (6, 9.730201),
        (7, 19.730201),
        (8, 39.300310),
        (9, 57.692098),
        (10, 67.220311),
    ], '-p', 'darkseagreen'),
    ('Front-Attacked-Rate(Safe)', [
        (0, 0),
        (1, 0),
        (2, 0),
        (3, 0),
        (4, 0.196490),
        (5, 0.424058),
        (6, 0.557805),
        (7, 0.611316),
        (8, 1.048701),
        (9, 3.589304),
        (10, 5.896310),
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
    plt.savefig('front-attack-rate-6w.pdf', format='pdf')
    plt.show()


if __name__ == '__main__':
    do_plot()
