import {
  LocalParticipant, Participant, RemoteParticipant,
} from 'livekit-client';
import {
  AudioRenderer, LiveKitRoom, ParticipantView, StageProps,
} from 'livekit-react';
import React, { useEffect } from 'react';
import {
  onConnected, stopRecording, TemplateProps, useParams,
} from './common';
import styles from './GridPage.module.css';

export default function GridPage({ interfaceStyle }: TemplateProps) {
  const { url, token } = useParams();

  if (!url || !token) {
    return <div className="error">missing required params url and token</div>;
  }

  let containerClass = 'roomContainer';
  if (interfaceStyle) {
    containerClass += ` ${interfaceStyle}`;
  }
  return (
    <div className={containerClass}>
      <LiveKitRoom
        url={url}
        token={token}
        onConnected={onConnected}
        onLeave={stopRecording}
        stageRenderer={renderStage}
        adaptiveVideo
      />
    </div>
  );
}

const renderStage: React.FC<StageProps> = ({ roomState }: StageProps) => {
  const {
    error, room, participants, audioTracks,
  } = roomState;
  const [visibleParticipants, setVisibleParticipants] = React.useState<Participant[]>([]);
  const [gridClass, setGridClass] = React.useState(styles.grid1x1);

  // select participants to display on first page, keeping ordering consistent if possible.
  useEffect(() => {
    // remove any participants that are no longer connected
    const newParticipants: Participant[] = [];
    visibleParticipants.forEach((p) => {
      if (room?.participants.has(p.sid)) {
        newParticipants.push(p);
      }
    });

    // ensure active speaker is visible
    room?.activeSpeakers.forEach((speaker) => {
      if (newParticipants.includes(speaker) || speaker instanceof LocalParticipant) {
        return;
      }
      newParticipants.unshift(speaker);
    });

    for (let i = 0; i < participants.length; i += 1) {
      const participant = participants[i];
      if (participant instanceof RemoteParticipant && !newParticipants.includes(participant)) {
        newParticipants.push(participants[i]);
      }
      // max of 6x6 grid
      if (newParticipants.length > 36) {
        return;
      }
    }
    if (newParticipants.length >= 36) {
      setGridClass(styles.grid6x6);
      newParticipants.splice(36, newParticipants.length - 36);
    } else if (newParticipants.length >= 31) {
      setGridClass(styles.grid6x6);
    } else if (newParticipants.length >= 26) {
      setGridClass(styles.grid6x5);
    } else if (newParticipants.length >= 21) {
      setGridClass(styles.grid5x5);
    } else if (newParticipants.length >= 17) {
      setGridClass(styles.grid5x4);
    } else if (newParticipants.length >= 13) {
      setGridClass(styles.grid4x4);
    } else if (newParticipants.length >= 10) {
      setGridClass(styles.grid4x3);
    } else if (newParticipants.length >= 7) {
      setGridClass(styles.grid3x3);
    } else if (newParticipants.length >= 5) {
      setGridClass(styles.grid3x2);
    } else if (newParticipants.length >= 3) {
      setGridClass(styles.grid2x2);
    } else if (newParticipants.length === 2) {
      setGridClass(styles.grid2x1);
    } else if (newParticipants.length === 1) {
      setGridClass(styles.grid1x1);
    }
    setVisibleParticipants(newParticipants);
  }, [participants]);

  if (error) {
    return <div className="error">{error}</div>;
  }

  if (!room) {
    return <div>room closed</div>;
  }

  if (visibleParticipants.length === 0) {
    return <div />;
  }

  const audioRenderers = audioTracks.map((track) => (
    <AudioRenderer key={track.sid} track={track} isLocal={false} />
  ));

  return (
    <div className={`${styles.stage} ${gridClass}`}>
      {visibleParticipants.map((participant) => (
        <ParticipantView
          key={participant.identity}
          participant={participant}
          orientation="landscape"
          width="100%"
          height="100%"
          adaptiveVideo
        />
      ))}
      {audioRenderers}
    </div>
  );
};
